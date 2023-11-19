package tools

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"image/draw"

	"github.com/fogleman/gg"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type LISTools struct {
	equipmentName string
	URL           string
	ipAddress     string
	port          int
	SAVE_RESULTS  bool
	toLog         bool
}

const (
	OFF      = 0
	ON       = 1
	ANALYZER = 0
	LIS      = 1
	EMPTY    = 2
)

func NewLISTools(equipmentName string) *LISTools {
	return &LISTools{
		equipmentName: equipmentName,
		URL:           "",
		ipAddress:     "",
		port:          0,
		SAVE_RESULTS:  false,
		toLog:         true,
	}
}

func (lt *LISTools) ErrorLog(errorMessage string, e error) {
	if lt.toLog {
		fmt.Println(errorMessage)
		lt.LogMessage(errorMessage, 2)
		// Additional error handling/logic for Go error handling
	}
}

func (lt *LISTools) LogMessage(message string, sender int) {
	messageOwner := ""
	date := time.Now().Format("2006-01-02 15:04:05")
	fileName := lt.GetFolderDir(1, "") + "/" + time.Now().Format("2006-01-02") + ".txt"

	switch sender {
	case 0:
		messageOwner = ": <--"
	case 1:
		messageOwner = ": -->"
	case 2:
		messageOwner = ": ---"
	}

	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		lt.ErrorLog(err.Error(), err)
		return
	}
	defer file.Close()

	_, err = file.WriteString("\n" + date + messageOwner + message)
	if err != nil {
		lt.ErrorLog(err.Error(), err)
		return
	}
}

func (lt *LISTools) GetFolderDir(choice int, eqName string) string {
	homeDir, _ := os.UserHomeDir()
	tempDir := filepath.Join(homeDir, "Documents")
	fileDir := filepath.Join(tempDir, "lis_logs")

	if _, err := os.Stat(fileDir); os.IsNotExist(err) {
		if err := os.Mkdir(fileDir, 0755); err != nil {
			fmt.Println("LIS logs not created")
		}
	}

	var equipmentPath string
	if choice == 2 {
		equipmentPath = filepath.Join(fileDir, eqName)
	} else {
		equipmentPath = filepath.Join(fileDir, lt.equipmentName)
	}

	lisModelDir := filepath.Join(equipmentPath)
	if _, err := os.Stat(lisModelDir); os.IsNotExist(err) {
		if err := os.Mkdir(lisModelDir, 0755); err != nil {
			fmt.Println(strings.ToUpper(lt.equipmentName) + " not created")
		}
	}

	ErrorLogDir := filepath.Join(lisModelDir, "err_logs")
	if _, err := os.Stat(ErrorLogDir); os.IsNotExist(err) {
		if err := os.Mkdir(ErrorLogDir, 0755); err != nil {
			fmt.Println("Error Logs not created")
		}
	}

	messageLogDir := filepath.Join(lisModelDir, "mess_logs")
	if _, err := os.Stat(messageLogDir); os.IsNotExist(err) {
		if err := os.Mkdir(messageLogDir, 0755); err != nil {
			fmt.Println("Message Logs not created")
		}
	}

	configDir := filepath.Join(lisModelDir, "configs")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.Mkdir(configDir, 0755); err != nil {
			fmt.Println("Config folder is not created")
		}
	}

	switch choice {
	case 0:
		return ErrorLogDir
	case 1:
		return messageLogDir
	case 2:
		return configDir
	default:
		return ""
	}
}
func (lt *LISTools) Parser(text, delimiter string) []string {
	return strings.Split(text, delimiter)
}
func (lt *LISTools) displayMessages(message string, sender int) {
	messageOwner := ""
	switch sender {
	case 0:
		messageOwner = "<--"
	case 1:
		messageOwner = "-->"
	}

	fmt.Println(messageOwner + message)
}

func (lt *LISTools) LogAndDisplayMessage(message string, sender int) {
	if lt.toLog {
		lt.LogMessage(message, sender)
		lt.displayMessages(message, sender)
	}
}

func (lt *LISTools) GetHistogramAsPNG(graphType, data string) string {
	// var image string
	width, height := 500, 500
	dc := gg.NewContext(width, height)

	// Calculate points using reverseAlgorithm function
	points := lt.reverseAlgorithm(data)

	x, xAxe, yAxe, y := 70.0, 50.0, 450.0, 450.0

	// Set up font for drawing text
	fontPath := "./luxisr.ttf" // Replace with the path to a TTF font file
	err := dc.LoadFontFace(fontPath, 24)
	if err != nil {
		log.Fatal(err)
	}

	// Draw graphType label
	dc.DrawStringAnchored(graphType, 50, 60, 0.5, 0.5)

	// Draw axes and labels
	dc.DrawLine(40, yAxe, 460, yAxe)
	dc.DrawLine(xAxe, 70, xAxe, 460)
	dc.DrawStringAnchored("0", xAxe-5, yAxe+30, 0.5, 0.5)

	// Set color based on graphType
	var graphColor color.Color
	switch graphType {
	case "PLT":
		graphColor = color.RGBA{0, 255, 0, 255} // Green
	case "RBC":
		graphColor = color.RGBA{255, 0, 0, 255} // Red
	default:
		graphColor = color.Black
	}

	// Draw graph points
	for i := 0; i < len(points)-1; i++ {
		a := x + 3
		dc.SetColor(graphColor)
		dc.DrawLine(x, y-float64(points[i]), a, y-float64(points[i+1]))
		dc.Stroke()
		x += 3
	}

	// Draw additional lines
	dc.SetColor(color.Black)
	dc.DrawLine(70, 70, 70, yAxe)
	dc.DrawLine(x, 70, x, yAxe)
	dc.Stroke()

	// Encode image as PNG and return as base64 string
	var buf bytes.Buffer
	err = png.Encode(&buf, dc.Image())
	if err != nil {
		log.Print(err)
		return ""
	}
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return encoded
}
func DrawText(dst draw.Image, text string, x, y int, fontFace font.Face, c color.Color) {
	d := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(c),
		Face: fontFace,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}
	d.DrawString(text)
}

func (lt *LISTools) CyrToLat(text string) string {
	cyr := []rune{' ', 'а', 'б', 'в', 'г', 'д', 'е', 'ё', 'ж', 'з', 'и', 'й', 'к', 'л', 'м', 'н', 'о', 'п', 'р', 'с', 'т', 'у', 'ф', 'х', 'ц', 'ч', 'ш', 'щ', 'ь', 'ы', 'ъ', 'э', 'ю', 'я', 'А', 'Б', 'В', 'Г', 'Д', 'Е', 'Ё', 'Ж', 'З', 'И', 'Й', 'К', 'Л', 'М', 'Н', 'О', 'П', 'Р', 'С', 'Т', 'У', 'Ф', 'Х', 'Ц', 'Ч', 'Ш', 'Щ', 'Ь', 'Ы', 'Ъ', 'Э', 'Ю', 'Я'}
	lat := []string{" ", "a", "b", "v", "g", "d", "e", "e", "zh", "z", "i", "y", "k", "l", "m", "n", "o", "p", "r", "s", "t", "u", "f", "h", "ts", "ch", "sh", "sch", "", "i", "", "e", "yu", "ya", "A", "B", "V", "G", "D", "E", "E", "Zh", "Z", "I", "Y", "K", "L", "M", "N", "O", "P", "R", "S", "T", "U", "F", "H", "Ts", "Ch", "Sh", "Sch", "", "I", "", "E", "Yu", "Ya"}
	var builder strings.Builder
	isCyr := false

	for _, char := range text {
		found := false
		for j, cyrChar := range cyr {
			if char == cyrChar {
				builder.WriteString(lat[j])
				isCyr = true
				found = true
				break
			}
		}
		if !found {
			builder.WriteRune(char)
		}
	}

	if isCyr {
		return builder.String()
	}
	return text
}

func (lt *LISTools) reverseAlgorithm(data string) []int {
	decText := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	counter := 1
	decHex := make([]int, 0)
	var binaryData strings.Builder
	started := false

	for i := 0; i < len(data); i++ {
		for j := 0; j < len(decText); j++ {
			if decText[j] == data[i] {
				// hex := strconv.FormatInt(int64(j), 16)
				binary := strconv.FormatInt(int64(j), 2)
				zeros := ""

				if binary == "0" {
					binary = "000000"
				} else {
					for k := 0; k < 6-len(binary); k++ {
						zeros += "0"
					}
					binary = zeros + binary
				}

				binaryData.WriteString(binary)

				if counter == 4 {
					for l := 0; l < 24; l += 8 {
						el := binaryData.String()[l : l+8]
						if el != "00000000" || started {
							point, _ := strconv.ParseInt(el, 2, 64)
							if int(point) > 2 {
								decHex = append(decHex, int(point))
							}
							started = true
						}
					}
					binaryData.Reset()
					counter = 1
				} else {
					counter++
				}
			}
		}
	}
	return decHex
}

func (lt *LISTools) GetScattergramAsPNG(graphType, data string) string {
	width, height := 500, 500
	dc := gg.NewContext(width, height)

	// Calculate points using reverseAlgorithm function
	points := lt.reverseAlgorithm(data)

	_, xAxe, yAxe, _ := 70.0, 20.0, 480.0, 450.0

	// Set up font for drawing text
	fontPath := "./luxisr.ttf" // Replace with the path to a TTF font file
	err := dc.LoadFontFace(fontPath, 24)
	if err != nil {
		log.Fatal(err)
	}

	// Draw graphType label
	dc.DrawStringAnchored(graphType, 50, 60, 0.5, 0.5)

	// Draw axes
	dc.DrawLine(20, yAxe, 460, yAxe)
	dc.DrawLine(xAxe, 70, xAxe, 480)

	// Set color based on graphType
	if graphType == "BASO" {
		dc.SetColor(color.RGBA{255, 0, 0, 255}) // Red
	} else {
		dc.SetColor(color.RGBA{0, 0, 255, 255}) // Blue
	}

	// Draw scatter points
	for i := 0; i < len(points); i += 4 {
		dc.DrawPoint(xAxe+float64(points[i]), yAxe-float64(points[i+1]), 1)
	}

	// Encode image as PNG and return as base64 string
	var buf bytes.Buffer
	err = png.Encode(&buf, dc.Image())
	if err != nil {
		log.Print(err)
		return ""
	}
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return encoded
}
func (lt *LISTools) EnableLog(toLog bool) {
	lt.toLog = toLog
}
