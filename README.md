<h3 align="center">GhostPing</h3>
<h1 align="center"> <img src="https://github.com/yourusername/ghostping/blob/main/ghostping.png" alt="ghostping" width="500px"></h1>

GhostPing is a lightweight Go server that collects client location data (latitude, longitude, IP), enriches it with geolocation details, and securely forwards it to a Telegram chat for real-time monitoring.

# Download

```sh
git clone [https://github.com/yourusername/ghostping](https://github.com/yourusername/ghostping) && cd ghostping
```

# Usage

```sh
\$ go run main.go \[OPTIONS]
Options:
-h, --help        Show this help message
-token TOKEN      Telegram bot token (required)
-chat CHAT\_ID     Telegram chat ID (required)
-port PORT        Port to run the server (default: 8088)

Description:
GhostPing listens for POST requests on /send-location. It extracts latitude and longitude from the JSON body, resolves the client’s IP into city/country/ISP details, then pushes a formatted message to your Telegram bot. Useful for location logging and monitoring.
```

# Example

```sh
\$ go run main.go -token 123456\:ABC-DEF -chat 987654321 -port 8088
```

# API Example

```http
POST [http://localhost:8088/send-location](http://localhost:8088/send-location)
Content-Type: application/json

{
"latitude": "-22.9129",
"longitude": "-43.2003"
}
```