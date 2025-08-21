<h3 align="center">GhostPing</h3>
<h1 align="center"> <img src="https://github.com/k4rkarov/ghostping/blob/main/carbon.png" alt="ghostping" width="700px"></h1>

A Go-based server that collects client location data (latitude, longitude, IP), enriches it with geolocation details, and securely forwards it to a Telegram chat for real-time monitoring.

<br>

# Installation Instructions

`ghostping` requires **Go 1.18** or later to install successfully. Run the following command to install the latest version: 

```sh
go install github.com/k4rkarov/ghostping@latest
````

Ensure your Go environment is properly configured to fetch and install dependencies.

# Usage

```sh
ghostping -h
```

This will display the help menu.

```console

                                                                                      
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

Usage:
  ghostping [options]

Options:
  -token TOKEN     Telegram bot token (required)
  -chat CHAT_ID    Telegram chat ID (required)
  -port PORT       Port to run the server (default: 8088)
  -h, --help       Show this help message

Description:
  GhostPing listens for POST requests on /send-location. 
  It extracts latitude and longitude from the JSON body, 
  resolves the client’s IP into city/country/ISP details, 
  then pushes a formatted message to your Telegram bot.
```

# Running GhostPing

### Start the server

```bash
ghostping -token 123456:ABC-DEF -chat 987654321 -port 8088
Server running on port 8088...
```

### Send a location payload

```http
POST http://localhost:8088/send-location
Content-Type: application/json

{
  "latitude": "-22.9129",
  "longitude": "-43.2003"
}
```

### Telegram output

```text
📍 Location
Latitude: -22.9129
Longitude: -43.2003

🌐 Network
IP: 203.0.113.10
City: Rio de Janeiro
Country: Brazil
ISP: Claro NXT Telecomunicacoes Ltda

🖥 Device
User-Agent: Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Mobile Safari/537.36
Language: pt-BR
Timezone: America/Sao_Paulo
Screen: 412x915
DPR: 1.75
Platform: Linux armv81
Battery: 18% not charging
Connection: 4g, downlink: 10Mb/s, rtt: 150ms
CPU Cores: 8
Memory: 4.0 GB
Cookies Enabled: true
Plugins: 
Touch Points: 2

🔗 [View on Google Maps](https://www.google.com/maps?q=-22.9129,-43.2003)
```

# Notes

* You must create a Telegram Bot using [@BotFather](https://t.me/BotFather) and obtain both the **bot token** and the **chat ID**.
* The server is designed to be robust: it handles missing coordinates, unavailable IP data, and Telegram API errors gracefully.
* For production use, consider running GhostPing as a `systemd` service or inside a container.

<details>
<summary><strong>How to create your Telegram bot with BotFather</strong></summary>

1. Open Telegram and search for [@BotFather](https://t.me/BotFather).
2. Start a chat and send `/newbot`.
3. Follow the instructions: provide a bot name and username.
4. Once created, BotFather will give you a **bot token**. Save it.
5. To get your **chat ID**, start a chat with your bot and send a message, then use the following link in your browser:

   ```
   https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates
   ```

   Look for `"chat":{"id":<YOUR_CHAT_ID>}` in the JSON response. Use that ID in your GhostPing command.

</details>

<details>
<summary><strong>How to configure ngrok</strong></summary>

1. Download and install ngrok from [https://ngrok.com/download](https://ngrok.com/download).
2. Authenticate your account:

   ```bash
   ngrok authtoken <YOUR_NGROK_AUTH_TOKEN>
   ```
3. Start a tunnel pointing to your GhostPing server:

   ```bash
   ngrok http 8088
   ```
4. ngrok will give you a public URL like `https://abcd1234.ngrok-free.app/`.
   Share this URL with your users – when they open it in a browser, GhostPing will automatically request their location and send it to your Telegram bot.

</details>

For any issues or feature requests, please open an issue on the [GitHub repository](https://github.com/yourusername/ghostping).

Enjoy using GhostPing for real-time location tracking!