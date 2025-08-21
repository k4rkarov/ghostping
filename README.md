<h3 align="center">GhostPing</h3>
<h1 align="center"> <img src="https://github.com/k4rkarov/ghostping/blob/main/ghostping.png" alt="ghostping" width="400px"></h1>

A Go-based server that collects client location data (latitude, longitude, IP), enriches it with geolocation details, and securely forwards it to a Telegram chat for real-time monitoring.

<br>

# Installation Instructions

`ghostping` requires **Go 1.18** or later to install successfully. Run the following command to install the latest version: 

```sh
go install github.com/yourusername/ghostping/cmd/ghostping@latest
````

Ensure your Go environment is properly configured to fetch and install dependencies.

# Usage

```sh
ghostping -h
```

This will display the help menu.

```console

   ____ _               _   ____  _             
  / ___| |__   ___  ___| |_|  _ \(_)_ __   __ _ 
 | |  _| '_ \ / _ \/ __| __| |_) | | '_ \ / _` |
 | |_| | | | |  __/\__ \ |_|  __/| | | | | (_| |
  \____|_| |_|\___||___/\__|_|   |_|_| |_|\__, |
                                           |___/ 

       by yourusername (v1.0)

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

```
$ ghostping -token 123456:ABC-DEF -chat 987654321 -port 8088
Server running on port 8088...
```

### Send a location payload

```
POST http://localhost:8088/send-location
Content-Type: application/json

{
  "latitude": "-22.9129",
  "longitude": "-43.2003"
}
```

### Telegram output

```
📍 Location Received:
Latitude: -22.9129
Longitude: -43.2003

🌐 IP: 203.0.113.10
🏙 City: Rio de Janeiro
🏳 Country: Brazil
📡 ISP: Example Telecom

🔗 [View on Google Maps](https://www.google.com/maps?q=-22.9129,-43.2003)
```

# Notes

* You must create a Telegram Bot using [@BotFather](https://t.me/BotFather) and obtain both the **bot token** and the **chat ID**.
* The server is designed to be robust: it handles missing coordinates, unavailable IP data, and Telegram API errors gracefully.
* For production use, consider running GhostPing as a `systemd` service or inside a container.

For any issues or feature requests, please open an issue on the [GitHub repository](https://github.com/yourusername/ghostping).

Enjoy using GhostPing for real-time location tracking!