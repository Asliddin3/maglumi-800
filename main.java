
package equipments;
 
import org.json.JSONObject;
import java.util.Calendar;
import java.text.SimpleDateFormat;
import java.util.Iterator;
import java.io.Reader;
import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.URL;
import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;
import java.io.ByteArrayOutputStream;
import java.io.DataInputStream;
import java.io.IOException;
import java.net.ServerSocket;
import java.io.InputStream;
import java.io.DataOutputStream;
import java.net.Socket;
import tools.LISTools;
import java.util.ArrayList;
 
public class AsServerMaglumi800 implements BaseEquipment
{
    private int MESSAGE_ORDER;
    private int MESSAGE_LENGTH;
    private int MESSAGE_TYPE;
    private final int QUERY = 0;
    private final int RESULTS = 1;
    private ArrayList<String> dataReceived;
    private ArrayList<String> test_order;
    private ArrayList<String> analyzerMessage;
    public static LISTools tools;
    private Socket socket;
    private DataOutputStream os;
    private InputStream is;
    private boolean connectionStatus;
    private String eqName;
    private String host;
    private String apiURL;
    private String port;
    private String ipAddress;
    private ServerSocket server;
 
    public AsServerMaglumi800() {
        this.MESSAGE_ORDER = 0;
        this.MESSAGE_LENGTH = 0;
        this.MESSAGE_TYPE = 2;
        this.os = null;
        this.is = null;
        this.server = null;
    }
 
    public void setParams(final String name, final String host, final String ipAddress, final String port, final String apiURL) {
        this.eqName = name;
        this.host = host;
        this.apiURL = apiURL;
        this.ipAddress = ipAddress;
        this.port = port;
        AsServerMaglumi800.tools = new LISTools(this.eqName);
        this.dataReceived = new ArrayList<String>();
        this.test_order = new ArrayList<String>();
        this.analyzerMessage = new ArrayList<String>();
        try {
            if (this.server == null) {
                this.server = new ServerSocket(Integer.parseInt(port));
                AsServerMaglumi800.tools.logAndDisplayMessage(name + " created", 2);
            }
        }
        catch (IOException e) {
            AsServerMaglumi800.tools.errorLog("Unable to create server socket with port: " + port + " - " + this.eqName, (Exception)e);
        }
    }
 
    public boolean connect() {
        try {
            this.socket = this.server.accept();
            AsServerMaglumi800.tools.logAndDisplayMessage("Connected to " + this.ipAddress + ":" + this.port + " - " + this.eqName, 2);
            this.os = new DataOutputStream(this.socket.getOutputStream());
            this.is = new DataInputStream(this.socket.getInputStream());
            AsServerMaglumi800.tools.logAndDisplayMessage("\n--------------------------------------------------------------------------------------------------------------------------\n", 2);
        }
        catch (Exception e) {
            this.disconnect();
            AsServerMaglumi800.tools.errorLog("Unable to create connection to " + this.ipAddress + ":" + this.port + " - " + this.eqName, e);
        }
        return this.socket.isConnected();
    }
 
    public void start() {
        try {
            while (true) {
                final ByteArrayOutputStream baos = new ByteArrayOutputStream();
                final byte[] res_arr = new byte[64000];
                try {
                    baos.write(res_arr, 0, this.is.read(res_arr));
                }
                catch (NullPointerException ne) {
                    AsServerMaglumi800.tools.errorLog(ne.getMessage(), (Exception)ne);
                    break;
                }
                final byte[] result = baos.toByteArray();
                final StringBuilder charBuffer = new StringBuilder();
                for (final byte b : result) {
                    charBuffer.append((char)b);
                }
                for (String buffer = charBuffer.toString(); !buffer.isEmpty(); buffer = "") {
                    if (buffer.contains("\u0005")) {
                        AsServerMaglumi800.tools.logAndDisplayMessage("<ENQ>", 0);
                        this.os.writeBytes("\u0006");
                        AsServerMaglumi800.tools.logAndDisplayMessage("<ACK>", 1);
                    }
                    else if (buffer.contains("\u0002")) {
                        AsServerMaglumi800.tools.logAndDisplayMessage("<STX>", 0);
                        this.os.writeBytes("\u0006");
                        AsServerMaglumi800.tools.logAndDisplayMessage("<ACK>", 1);
                    }
                    else if (buffer.contains("\u0004")) {
                        AsServerMaglumi800.tools.logAndDisplayMessage("<EOT>", 0);
                        this.os.writeBytes("\u0006");
                        AsServerMaglumi800.tools.logAndDisplayMessage("<ACK>", 1);
                        if (!this.dataReceived.isEmpty()) {
                            this.parseReceived();
                            switch (this.MESSAGE_TYPE) {
                                case 0: {
                                    this.test_order = this.getOrder();
                                    if (!this.test_order.isEmpty()) {
                                        AsServerMaglumi800.tools.logAndDisplayMessage("Analyzer query was processed and test order was prepared", 2);
                                        this.MESSAGE_ORDER = 0;
                                        this.MESSAGE_LENGTH = this.test_order.size();
                                        this.os.writeBytes(this.test_order.get(this.MESSAGE_ORDER++));
                                        AsServerMaglumi800.tools.logAndDisplayMessage("<ENQ>", 1);
                                        break;
                                    }
                                    break;
                                }
                                case 1: {
                                    this.saveResults();
                                    AsServerMaglumi800.tools.logAndDisplayMessage("Analyzer test results were processed and saved", 2);
                                    break;
                                }
                                default: {
                                    AsServerMaglumi800.tools.logAndDisplayMessage("The message was empty", 2);
                                    break;
                                }
                            }
                        }
                        this.dataReceived = new ArrayList<String>();
                    }
                    else if (buffer.contains("\u0006")) {
                        AsServerMaglumi800.tools.logAndDisplayMessage("<ACK>", 0);
                        if (!this.test_order.isEmpty()) {
                            if (this.MESSAGE_ORDER == this.MESSAGE_LENGTH) {
                                this.MESSAGE_ORDER = 0;
                                this.MESSAGE_LENGTH = 0;
                                this.test_order = new ArrayList<String>();
                            }
                            else {
                                final String test = this.test_order.get(this.MESSAGE_ORDER++) + "\r\n";
                                this.os.writeBytes(test);
                                AsServerMaglumi800.tools.logAndDisplayMessage(test, 1);
                            }
                        }
                    }
                    else if (buffer.contains("\u0015")) {
                        AsServerMaglumi800.tools.logAndDisplayMessage("<NAK>", 0);
                        this.os.writeBytes("\u0006");
                        AsServerMaglumi800.tools.logAndDisplayMessage("<ACK>", 1);
                    }
                    else {
                        AsServerMaglumi800.tools.logAndDisplayMessage(buffer, 0);
                        this.dataReceived.add(buffer);
                    }
                }
            }
        }
        catch (Exception ex) {
            this.disconnect();
            AsServerMaglumi800.tools.errorLog("Error while reading from and writing to stream: " + this.eqName, ex);
        }
    }
 
    public void disconnect() {
        try {
            this.setConnectionStatus(false);
            if (this.os != null) {
                this.os.close();
            }
            if (this.is != null) {
                this.is.close();
            }
            if (this.socket != null && !this.socket.isClosed()) {
                this.socket.close();
            }
            AsServerMaglumi800.tools.logAndDisplayMessage(this.eqName + " disconnected", 2);
        }
        catch (Exception e) {
            AsServerMaglumi800.tools.errorLog("Error while closing streams and connection: " + this.eqName, e);
        }
    }
 
    public void parseReceived() {
        try {
            switch (this.dataReceived.get(1).charAt(0)) {
                case 'Q': {
                    this.MESSAGE_TYPE = 0;
                    break;
                }
                case 'P': {
                    this.MESSAGE_TYPE = 1;
                    break;
                }
                default: {
                    this.MESSAGE_TYPE = 2;
                    break;
                }
            }
            this.analyzerMessage = this.dataReceived;
        }
        catch (Exception e) {
            this.MESSAGE_TYPE = 2;
            AsServerMaglumi800.tools.errorLog("Error in parsing the received message and the message is not saved: " + this.eqName, e);
        }
        finally {
            this.dataReceived = new ArrayList<String>();
        }
    }
 
    public void saveResults() {
        String barcode = "";
        final ArrayList<String[]> res_arr = new ArrayList<String[]>();
        for (final String el : this.analyzerMessage) {
            switch (el.charAt(0)) {
                case 'H': {}
                case 'P': {
                    continue;
                }
                case 'O': {
                    barcode = AsServerMaglumi800.tools.parser(el, "|")[2];
                    AsServerMaglumi800.tools.logAndDisplayMessage("Sample Barcode: " + barcode, 2);
                    continue;
                }
                case 'R': {
                    res_arr.add(AsServerMaglumi800.tools.parser(el, "|"));
                }
                case 'L': {
                    continue;
                }
            }
        }
        final ArrayList<String> params_arr = new ArrayList<String>();
        try {
            final String head_barcode = "method=apiResultSave&lisResult[name]=" + URLEncoder.encode(this.eqName, StandardCharsets.UTF_8.toString()) + "&lisResult[host]=" + URLEncoder.encode(this.host, StandardCharsets.UTF_8.toString()) + "&lisResult[barcode]=" + URLEncoder.encode(barcode, StandardCharsets.UTF_8.toString());
            for (final String[] el2 : res_arr) {
                final String code = AsServerMaglumi800.tools.parser(el2[2], "^")[3];
                final String params = "&lisResult[code]=" + URLEncoder.encode(code, StandardCharsets.UTF_8.toString()) + "&lisResult[R][res]=" + URLEncoder.encode(el2[3], StandardCharsets.UTF_8.toString()) + "&lisResult[R][unit]=" + URLEncoder.encode(el2[4], StandardCharsets.UTF_8.toString()) + "&lisResult[R][norms]=" + URLEncoder.encode(el2[5], StandardCharsets.UTF_8.toString()) + "&lisResult[R][flag]=" + URLEncoder.encode(el2[6], StandardCharsets.UTF_8.toString());
                params_arr.add(head_barcode + params);
            }
            AsServerMaglumi800.tools.logAndDisplayMessage("Sending results to API...: " + this.eqName, 2);
            for (final String params2 : params_arr) {
                final URL curl = new URL(this.apiURL);
                final HttpURLConnection conn = (HttpURLConnection)curl.openConnection();
                conn.setRequestMethod("POST");
                conn.setRequestProperty("User-Agent", "Mozilla/5.0");
                conn.setRequestProperty("Accept-Language", "en-US,en;q=0.5");
                conn.setDoOutput(true);
                try (final DataOutputStream wr = new DataOutputStream(conn.getOutputStream())) {
                    wr.writeBytes(params2);
                    wr.flush();
                }
                AsServerMaglumi800.tools.logAndDisplayMessage("\nSending 'POST' request to URL : " + AsServerMaglumi800.tools.URL, 2);
                AsServerMaglumi800.tools.logAndDisplayMessage("Post parameters : " + params2, 2);
                AsServerMaglumi800.tools.logAndDisplayMessage("API Response Code : " + conn.getResponseCode(), 2);
                try (final BufferedReader in = new BufferedReader(new InputStreamReader(conn.getInputStream()))) {
                    final StringBuilder response = new StringBuilder();
                    String line;
                    while ((line = in.readLine()) != null) {
                        response.append(line);
                    }
                    AsServerMaglumi800.tools.logAndDisplayMessage("API Response: " + response.toString(), 2);
                }
            }
        }
        catch (Exception e) {
            AsServerMaglumi800.tools.errorLog("Error while sending results to API: " + this.eqName, e);
        }
    }
 
    public ArrayList<String> getOrder() {
        ArrayList<String> order = new ArrayList<String>();
        final String[] barcodeArr = AsServerMaglumi800.tools.parser(AsServerMaglumi800.tools.parser((String)this.analyzerMessage.get(1), "|")[2], "^");
        String barcode = "00000000";
        if (barcodeArr.length > 1) {
            barcode = barcodeArr[1];
        }
        AsServerMaglumi800.tools.logAndDisplayMessage("Sample Barcode: " + barcode, 2);
        try {
            final String params = "method=apiOrderGet&order[name]=" + URLEncoder.encode(this.eqName, StandardCharsets.UTF_8.toString()) + "&order[host]=" + URLEncoder.encode(this.host, StandardCharsets.UTF_8.toString()) + "&order[barcode]=" + URLEncoder.encode(barcode, StandardCharsets.UTF_8.toString());
            final URL curl = new URL(this.apiURL);
            final HttpURLConnection conn = (HttpURLConnection)curl.openConnection();
            conn.setRequestMethod("POST");
            conn.setRequestProperty("User-Agent", "Mozilla/5.0");
            conn.setRequestProperty("Accept-Language", "en-US,en;q=0.5");
            conn.setDoOutput(true);
            AsServerMaglumi800.tools.logAndDisplayMessage("Getting orders from API...: " + this.eqName, 2);
            try (final DataOutputStream wr = new DataOutputStream(conn.getOutputStream())) {
                wr.writeBytes(params);
                wr.flush();
            }
            AsServerMaglumi800.tools.logAndDisplayMessage("\nSending 'POST' request to URL : " + AsServerMaglumi800.tools.URL, 2);
            AsServerMaglumi800.tools.logAndDisplayMessage("Post parameters : " + params, 2);
            AsServerMaglumi800.tools.logAndDisplayMessage("API Response Code : " + conn.getResponseCode(), 2);
            try (final BufferedReader in = new BufferedReader(new InputStreamReader(conn.getInputStream()))) {
                final StringBuilder response = new StringBuilder();
                String line;
                while ((line = in.readLine()) != null) {
                    response.append(line);
                }
                AsServerMaglumi800.tools.logAndDisplayMessage("API Response: " + response.toString(), 2);
                final String date = new SimpleDateFormat("yyyyMMdd").format(Calendar.getInstance().getTime());
                order.add("\u0005");
                order.add("\u0002");
                order.add("H|\\^&||PSWD|Maglumi800|||||Lis||P|E1394-97|" + date);
                order.add("P|1");
                if (!response.toString().equals("Order not found")) {
                    final JSONObject object = new JSONObject(response.toString());
                    for (int i = 0; i < object.length(); ++i) {
                        final JSONObject obj = object.getJSONObject(String.valueOf(i));
                        order.add("O|" + (i + 1) + "|" + barcode + "||^^^" + obj.get("code") + "|R");
                    }
                }
                order.add("L|1|N");
                order.add("\u0003");
                order.add("\u0004");
            }
        }
        catch (Exception e) {
            order = new ArrayList<String>();
            AsServerMaglumi800.tools.errorLog("Error while getting orders from API: " + this.eqName, e);
        }
        return order;
    }
 
    public ArrayList<ArrayList<String>> getOrderList() {
        return null;
    }
 
    public void setConnectionStatus(final boolean value) {
        this.connectionStatus = value;
    }
 
    public boolean getConnectionStatus() {
        return this.connectionStatus;
    }
 
    public String getType() {
        return "server";
    }
}