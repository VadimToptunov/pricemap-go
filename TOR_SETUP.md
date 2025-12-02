# Tor Setup for Bypassing Blocks

This project supports using Tor proxy to bypass IP-based blocking when scraping real estate data.

## Installation

### 1. Install Tor

**macOS:**
```bash
brew install tor
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt-get install tor
```

**Linux (CentOS/RHEL):**
```bash
sudo yum install tor
```

**Windows:**
Download and install from https://www.torproject.org/download/

### 2. Start Tor Service

**macOS/Linux:**
```bash
# Start Tor service
tor

# Or as a service
sudo systemctl start tor
sudo systemctl enable tor  # Enable on boot
```

**Windows:**
Start Tor Browser or run Tor as a service.

### 3. Verify Tor is Running

Check if Tor is listening on default ports:
```bash
# Check SOCKS5 proxy (default: 9050)
netstat -an | grep 9050

# Check control port (default: 9051)
netstat -an | grep 9051
```

## Configuration

### Environment Variables

Add these to your `.env` file or set as environment variables:

```bash
# Enable Tor proxy
USE_TOR=true

# Tor proxy settings (defaults shown)
TOR_PROXY_HOST=127.0.0.1
TOR_PROXY_PORT=9050
TOR_CONTROL_PORT=9051

# Optional: Tor control password (if set in torrc)
TOR_CONTROL_PASSWORD=your_password
```

### Docker Setup

If running in Docker, you need to:

1. **Option 1: Use host network** (simplest)
   ```yaml
   services:
     scraper:
       network_mode: "host"
       environment:
         USE_TOR: "true"
   ```

2. **Option 2: Run Tor in separate container**
   ```yaml
   services:
     tor:
       image: dperson/torproxy
       ports:
         - "9050:9050"
         - "9051:9051"
   
     scraper:
       depends_on:
         - tor
       environment:
         USE_TOR: "true"
         TOR_PROXY_HOST: "tor"
   ```

## Usage

### Automatic Circuit Rotation

The Tor control interface allows rotating circuits (changing IP addresses):

```go
import "pricemap-go/utils"

torControl := utils.NewTorControl()
err := torControl.RenewCircuit()
if err != nil {
    log.Printf("Failed to renew circuit: %v", err)
}
```

### Manual Rotation

You can manually rotate Tor circuits using the control port:

```bash
# Connect to Tor control port
telnet 127.0.0.1 9051

# Authenticate (if password is set)
AUTHENTICATE "your_password"

# Get new circuit (new IP)
SIGNAL NEWNYM

# Check status
GETINFO status/circuit-established
```

## Security Notes

1. **Tor Control Port**: By default, Tor control port only accepts connections from localhost. This is secure for local use.

2. **Password Protection**: For production, set a password in `torrc`:
   ```
   HashedControlPassword 16:872860B76453A77D60CA2BB8C1A7042072093276A3D701AD684053EC4C
   ```

3. **Rate Limiting**: Even with Tor, respect rate limits to avoid overloading servers.

4. **Legal Compliance**: Ensure your use of Tor complies with local laws and website terms of service.

## Troubleshooting

### Tor Not Connecting

1. Check if Tor is running:
   ```bash
   ps aux | grep tor
   ```

2. Check Tor logs:
   ```bash
   tail -f /var/log/tor/log
   # or
   journalctl -u tor -f
   ```

3. Test Tor connection:
   ```bash
   curl --socks5-hostname 127.0.0.1:9050 https://check.torproject.org/api/ip
   ```

### Slow Performance

Tor adds latency (typically 1-3 seconds). This is normal. Consider:
- Using Tor only for blocked sites
- Increasing request timeouts
- Running multiple Tor instances for parallel requests

### Connection Errors

If you see connection errors:
1. Verify Tor proxy is accessible
2. Check firewall settings
3. Ensure Tor has enough circuits established
4. Try renewing the circuit

## Testing

Test Tor integration:

```bash
# Set environment
export USE_TOR=true
export TOR_PROXY_HOST=127.0.0.1
export TOR_PROXY_PORT=9050

# Run scraper
go run cmd/scraper/main.go
```

## Performance

- **Without Tor**: ~100-500ms per request
- **With Tor**: ~1-3 seconds per request
- **Circuit Rotation**: ~2-5 seconds to establish new circuit

Consider using Tor selectively:
- Use Tor for sites that block you
- Use direct connection for open data APIs
- Rotate circuits periodically (every 10-50 requests)

