package main

import (
	"encoding/json"
	"fmt"
	"log"
	"maglumi800/tools"
	"net"
	"os"
	"strings"
	"time"
)

type AsServerMaglumi800 struct {
	MESSAGE_ORDER    int
	MESSAGE_LENGTH   int
	MESSAGE_TYPE     int
	QUERY            int
	RESULTS          int
	dataReceived     []string
	testOrder        []string
	analyzerMessage  []string
	tools            *tools.LISTools
	socket           net.Conn
	connectionStatus bool
	eqName           string
	host             string
	apiURL           string
	port             string
	ipAddress        string
	server           net.Listener
	log              log.Logger
}

func NewAsServerMaglumi800() *AsServerMaglumi800 {
	return &AsServerMaglumi800{
		MESSAGE_ORDER:  0,
		MESSAGE_LENGTH: 0,
		MESSAGE_TYPE:   2,
		QUERY:          0,
		RESULTS:        1,
		tools:          nil,
		socket:         nil,
		server:         nil,
	}
}

func (as *AsServerMaglumi800) SetParams(name, host, ipAddress, port, apiURL string) {
	as.eqName = name
	as.host = host
	as.apiURL = apiURL
	as.ipAddress = ipAddress
	as.port = port
	as.tools = tools.NewLISTools(as.eqName)
	as.dataReceived = make([]string, 0)
	as.testOrder = make([]string, 0)
	as.analyzerMessage = make([]string, 0)

	if as.server == nil {
		serverAddr := ipAddress + ":" + port
		server, err := net.Listen("tcp", serverAddr)
		if err != nil {
			as.tools.ErrorLog("Unable to create server socket with port: "+port+" - "+as.eqName, err)
			return
		}
		as.server = server
		as.tools.LogAndDisplayMessage(name+" created", 2)
	}
}
func (as *AsServerMaglumi800) Connect() bool {
	conn, err := as.server.Accept()
	if err != nil {
		as.socket.Close()
		as.tools.ErrorLog("Unable to create connection to "+as.ipAddress+":"+as.port+" - "+as.eqName, err)
		return false
	}

	as.socket = conn
	as.tools.LogAndDisplayMessage("Connected to "+as.ipAddress+":"+as.port+" - "+as.eqName, 2)

	as.socket = conn
	as.socket = conn
	as.tools.LogAndDisplayMessage("\n--------------------------------------------------------------------------------------------------------------------------\n", 2)

	return true
}

func (as *AsServerMaglumi800) Start() {
	for {
		fmt.Println("start connection")
		resArr := make([]byte, 64000)
		n, err := as.socket.Read(resArr)
		if err != nil {
			as.tools.ErrorLog(err.Error(), err)
			break
		}
		log.Println("gotten request from maglumi")
		result := resArr[:n]
		var charBuffer strings.Builder
		for _, b := range result {
			charBuffer.WriteByte(b)
		}
		buffer := charBuffer.String()

		for len(buffer) > 0 {
			switch {
			case strings.Contains(buffer, "\u0005"):
				log.Println("---->ENQ")
				log.Println("<----ACK")
				as.tools.LogAndDisplayMessage("<ENQ>", 0)
				as.socket.Write([]byte("\u0006"))
				as.tools.LogAndDisplayMessage("<ACK>", 1)

			case strings.Contains(buffer, "\u0002"):
				log.Println("---->STX")
				log.Println("<----ACK")
				as.tools.LogAndDisplayMessage("<STX>", 0)
				as.socket.Write([]byte("\u0006"))
				as.tools.LogAndDisplayMessage("<ACK>", 1)

			case strings.Contains(buffer, "\u0004"):
				log.Println("---->EOT")
				log.Println("<----ACK")
				as.tools.LogAndDisplayMessage("<EOT>", 0)
				as.socket.Write([]byte("\u0006"))
				as.tools.LogAndDisplayMessage("<ACK>", 1)
				if len(as.dataReceived) > 0 {
					log.Println("data received ", as.dataReceived)
					as.parseReceived()
					switch as.MESSAGE_TYPE {
					case 0:
						as.testOrder = as.getOrder()
						if len(as.testOrder) > 0 {
							log.Println("Start analyzer query ")
							as.tools.LogAndDisplayMessage("Analyzer query was processed and test order was prepared", 2)
							as.MESSAGE_ORDER = 0
							as.MESSAGE_LENGTH = len(as.testOrder)
							as.socket.Write([]byte(as.testOrder[as.MESSAGE_ORDER]))
							as.tools.LogAndDisplayMessage("<ENQ>", 1)
						}
					case 1:
						log.Println("Save analyzer results ")
						as.saveResults()
						as.tools.LogAndDisplayMessage("Analyzer test results were processed and saved", 2)
					default:
						as.tools.LogAndDisplayMessage("The message was empty", 2)
					}
					as.testOrder = nil
				}
				as.dataReceived = nil

			case strings.Contains(buffer, "\u0006"):
				log.Println("----->ACK")
				as.tools.LogAndDisplayMessage("<ACK>", 0)
				if len(as.testOrder) > 0 {
					if as.MESSAGE_ORDER == as.MESSAGE_LENGTH {
						as.MESSAGE_ORDER = 0
						as.MESSAGE_LENGTH = 0
						as.testOrder = nil
					} else {
						test := as.testOrder[as.MESSAGE_ORDER] + "\r\n"
						as.socket.Write([]byte(test))
						as.tools.LogAndDisplayMessage(test, 1)
					}
				}

			case strings.Contains(buffer, "\u0015"):
				log.Println("---->NAK")
				log.Println("<----ACK")
				as.tools.LogAndDisplayMessage("<NAK>", 0)
				as.socket.Write([]byte("\u0006"))
				as.tools.LogAndDisplayMessage("<ACK>", 1)

			default:
				as.tools.LogAndDisplayMessage(buffer, 0)
				as.dataReceived = append(as.dataReceived, buffer)
			}

			buffer = ""
		}
	}
}
func (as *AsServerMaglumi800) Disconnect() {
	as.SetConnectionStatus(false)

	// if as.os != nil {
	// 	as.os.Close()
	// }
	// if as.is != nil {
	// 	as.is.Close()
	// }
	if as.socket != nil {
		as.socket.Close()
	}

	as.tools.LogAndDisplayMessage(as.eqName+" disconnected", 2)
}

func (as *AsServerMaglumi800) parseReceived() {
	defer func() {
		as.dataReceived = nil
	}()

	if len(as.dataReceived) > 1 {
		switch as.dataReceived[1][0] {
		case 'Q':
			as.MESSAGE_TYPE = 0
		case 'P':
			as.MESSAGE_TYPE = 1
		default:
			as.MESSAGE_TYPE = 2
		}
		as.analyzerMessage = as.dataReceived
	} else {
		as.MESSAGE_TYPE = 2
		as.tools.ErrorLog("Error in parsing the received message and the message is not saved: "+as.eqName, nil)
	}
}

type Resulsts struct {
	Res  string `json:"res"`
	Code string `json:"code"`
	Unit string `json:"unit"`
	Norm string `json:"norm"`
	Flag string `json:"flag"`
}

func (as *AsServerMaglumi800) saveResults() {
	barcode := ""
	resArr := make([][]string, 0)
	for _, el := range as.analyzerMessage {
		switch el[0] {
		case 'H', 'P':
			continue
		case 'O':
			parts := strings.Split(el, "|")
			barcode = parts[2]
			as.tools.LogAndDisplayMessage("Sample Barcode: "+barcode, 2)
		case 'R':
			resArr = append(resArr, strings.Split(el, "|"))
		case 'L':
			continue
		}
	}

	// headBarcode := "method=apiResultSave&lisResult[name]=" + url.QueryEscape(as.eqName) +
	// 	"&lisResult[host]=" + url.QueryEscape(as.host) +
	// 	"&lisResult[barcode]=" + url.QueryEscape(barcode)
	results := make([]Resulsts, len(resArr))
	for _, el := range resArr {
		codeParts := strings.Split(el[2], "^")
		code := codeParts[3]
		results = append(results, Resulsts{
			Code: code,
			Res:  el[3],
			Unit: el[4],
			Norm: el[5],
			Flag: el[6],
		})
		// params := "&lisResult[code]=" + url.QueryEscape(code) +
		// 	"&lisResult[R][res]=" + url.QueryEscape(el[3]) +
		// 	"&lisResult[R][unit]=" + url.QueryEscape(el[4]) +
		// 	"&lisResult[R][norms]=" + url.QueryEscape(el[5]) +
		// 	"&lisResult[R][flag]=" + url.QueryEscape(el[6])
		// paramsArr = append(paramsArr, headBarcode+params)
	}
	res, err := json.MarshalIndent(&results, "", " ")
	if err != nil {
		fmt.Println("failed to parse")
	}
	fmt.Println("Gotten results", string(res))
}
func (as *AsServerMaglumi800) getOrder() []string {
	var order []string
	barcodeArr := as.tools.Parser(as.tools.Parser(as.analyzerMessage[1], "|")[2], "^")
	barcode := "00000000"
	if len(barcodeArr) > 1 {
		barcode = barcodeArr[1]
	}
	exists := true
	as.tools.LogAndDisplayMessage("Sample Barcode: "+barcode, 2)
	as.tools.LogAndDisplayMessage("Getting orders from API...: "+as.eqName, 2)

	date := time.Now().Format("20060102")
	order = append(order, "\u0005", "\u0002", "H|\\^&||PSWD|Maglumi800|||||Lis||P|E1394-97|"+date, "P|1")
	//Write there result codes
	codes := []string{""}
	if exists {

		for i, obj := range codes {
			order = append(order, fmt.Sprintf("O|%d|%s||^^^%s|R", i+1, barcode, obj))
		}
	}
	order = append(order, "L|1|N", "\u0003", "\u0004")

	return order
}

func (as *AsServerMaglumi800) getOrderList() [][]string {
	return nil
}

func (as *AsServerMaglumi800) SetConnectionStatus(value bool) {
	as.connectionStatus = value
}

func (as *AsServerMaglumi800) GetConnectionStatus() bool {
	return as.connectionStatus
}

func (as *AsServerMaglumi800) GetType() string {
	return "server"
}
func handleConnection(conn net.Conn, server *AsServerMaglumi800) {
	// Implement your connection handling logic here using the 'server' instance
	defer conn.Close()
	fmt.Println("server start")

	// Connect to the server
	// server.Connect()
	// Start processing data
	server.Start()
}

func main() {

	// Create an instance of AsServerMaglumi800
	logFile, err := os.Create("logfile.txt")
	if err != nil {
		fmt.Println("Error creating log file:", err)
		return
	}
	defer logFile.Close()
	// Create a TCP listener
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error creating listener:", err)
		return
	}
	defer listener.Close()
	log.SetOutput(logFile)
	fmt.Println("Server listening on", listener.Addr())
	server := AsServerMaglumi800{
		// Initialize your AsServerMaglumi800 properties here
		tools:  tools.NewLISTools("test"),
		apiURL: "localhost:8000",
		host:   "localhost",
		eqName: "Maglumi800",
		server: listener,
	}

	// Accept incoming connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		fmt.Println("gotten connection")
		server.socket = conn
		go handleConnection(conn, &server)
	}
}

// Implement other methods of BaseEquipment interface if needed
