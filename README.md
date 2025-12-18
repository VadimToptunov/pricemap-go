# PriceMap Go

[![CI](https://github.com/VadimToptunov/pricemap-go/workflows/CI/badge.svg)](https://github.com/VadimToptunov/pricemap-go/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/VadimToptunov/pricemap-go)](https://goreportcard.com/report/github.com/VadimToptunov/pricemap-go)
[![codecov](https://codecov.io/gh/VadimToptunov/pricemap-go/branch/main/graph/badge.svg)](https://codecov.io/gh/VadimToptunov/pricemap-go)
[![Go Version](https://img.shields.io/github/go-mod/go-version/VadimToptunov/pricemap-go)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Version 2.0** - Production Ready ✅

A sophisticated real estate price scraping and analysis system with global coverage, Tor integration, and advanced anti-blocking mechanisms.

📖 **[Complete Documentation](COMPREHENSIVE_GUIDE.md)** | 🚀 **[Quick Start](#quick-start)** | 🔧 **[Configuration](#configuration)** | ☸️ **[Kubernetes](k8s/README.md)**

## ✨ Key Features

- 🌍 **Global Coverage**: 100+ cities across Russia, UK, USA, Spain + Open Data portals
- 🔒 **Tor Integration**: Built-in anonymity with automatic IP rotation (enabled by default)
- 🛡️ **Anti-Blocking**: Retry logic, proxy pool, user-agent rotation, rate limiting
- 📊 **Factor Analysis**: Crime, transport, education, infrastructure scores
- 🗺️ **Free Maps**: OpenStreetMap/Leaflet (no API keys required!)
- ⚡ **High Performance**: Exponential backoff, caching, concurrent scraping
- 🐳 **Docker Ready**: One command to start everything
- 🧪 **Well Tested**: Comprehensive test suite

### New in v2.0

- ✅ Tor proxy with circuit rotation every 10 requests
- ✅ Proxy pool support (HTTP/HTTPS/SOCKS5)
- ✅ 10 rotating User-Agent strings
- ✅ Exponential backoff retry (3 attempts default)
- ✅ Random delays between requests (3-5 sec)
- ✅ Context cancellation support
- ✅ Improved error handling
- ✅ Production-ready configuration

## 🚀 Quick Start

### Docker (Recommended)

```bash
# 1. Clone repository
git clone https://github.com/VadimToptunov/pricemap-go.git
cd pricemap-go

# 2. Create config (optional - has good defaults)
cp env.example .env

# 3. Start everything (includes Tor!)
docker-compose up -d

# 4. Check status
docker-compose ps

# 5. View logs
docker-compose logs -f scraper
```

**That's it!** 🎉

- API: http://localhost:3000/api/v1/properties
- Frontend: Open `web/index.html` in browser
- Tor is enabled by default

### Local Development

```bash
# Install Go 1.21+
go version

# Start PostgreSQL
docker-compose up postgres -d

# Run scraper
go run cmd/scraper/main.go

# Run API server
go run cmd/server/main.go
```

## 📋 Project Structure

```
pricemap-go/
├── cmd/           # Entry points (server, scraper, scheduler)
├── api/           # HTTP handlers, middleware, routing
├── models/        # Data models
├── parsers/       # Website parsers with anti-blocking
├── services/      # Business logic (scraping, factors, metrics)
├── utils/         # Helpers (Tor, proxy pool, user-agents)
├── database/      # Database connection & migrations
├── config/        # Configuration management
└── web/           # Frontend (HTML/CSS/JS)
```

## 🌐 Supported Data Sources

### Commercial Sites (with anti-blocking)
- **Cian.ru** - Russia (30+ cities)
- **Rightmove.co.uk** - UK (25+ cities)
- **Zillow.com** - USA (30+ cities)
- **Idealista.com** - Spain (20+ cities)

### Open Data Portals (no API keys needed)
- NYC, London, Berlin, Paris, Tokyo, Sydney, Moscow

**Total:** 100+ cities, 4 countries, 200+ search combinations

## ⚙️ Configuration

All configuration via environment variables. See [`env.example`](env.example):

```bash
# Tor (enabled by default)
USE_TOR=true
TOR_PROXY_HOST=tor
TOR_PROXY_PORT=9050

# Rate Limiting
RATE_LIMIT_DELAY=3    # seconds between requests
MAX_RETRIES=3         # retry attempts
RETRY_DELAY=5         # exponential backoff base

# No API keys required! (OpenStreetMap + Nominatim are free)
GOOGLE_MAPS_API_KEY=  # NOT USED
OPENCAGE_API_KEY=     # OPTIONAL
```

📖 **[Full Configuration Guide](COMPREHENSIVE_GUIDE.md#configuration)**

## 🔥 Usage

### Using Docker

```bash
# Start all services
docker-compose up -d

# Start specific service
docker-compose up scraper -d

# View logs
docker-compose logs -f scraper
docker-compose logs -f tor

# Stop services
docker-compose down
```

### Using Makefile

```bash
make build          # Build all binaries
make run            # Run API server
make scrape         # Run scraper once
make schedule       # Run scheduler
make test           # Run tests
make docker-up      # Start Docker services
```

## 🔌 API Endpoints

Base URL: `http://localhost:3000/api/v1`

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/properties` | GET | List all properties (with filters) |
| `/properties/:id` | GET | Get property details |
| `/heatmap` | GET | Get heatmap data |
| `/stats` | GET | Get statistics |
| `/metrics` | GET | Get system metrics |
| `/metrics/parser/:parser` | GET | Get parser-specific metrics |

**Example:**
```bash
# Get properties in Moscow under $500k
curl "http://localhost:3000/api/v1/properties?city=Moscow&price_max=500000&limit=10"

# Get statistics
curl "http://localhost:3000/api/v1/stats"
```

📖 **[Full API Reference](COMPREHENSIVE_GUIDE.md#api-reference)**

## 🛡️ Anti-Blocking Features

### Tor Integration (Enabled by Default)
- Automatic circuit rotation every 10 requests
- Manual rotation on HTTP errors (429, 5xx)
- SOCKS5 proxy on port 9050
- Control port integration (9051)

### Proxy Pool
- Support for HTTP/HTTPS/SOCKS5
- Round-robin and random selection
- Automatic health checking
- Failed proxy removal

### Request Randomization
- 10 different User-Agent strings
- Random delays (3-5 seconds default)
- Realistic browser headers
- Exponential backoff on failures

### Retry Logic
```
Attempt 1: immediate
Attempt 2: wait 5 seconds
Attempt 3: wait 10 seconds
Attempt 4: wait 20 seconds
```

📖 **[Complete Anti-Blocking Guide](COMPREHENSIVE_GUIDE.md#anti-blocking-mechanisms)**

## 🧪 Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package
go test ./utils/... -v
```

**Test Results:** 23/23 passed ✅

## 📊 Data Analysis Features

The system analyzes multiple factors affecting property prices:

1. **Crime & Safety** (0-100)
   - Police API integration
   - Crime statistics by area
   
2. **Transportation** (0-100)
   - Metro/bus stop proximity
   - GTFS data parsing
   - Time to city center

3. **Education** (0-100)
   - School ratings
   - University proximity

4. **Infrastructure** (0-100)
   - Shops, parks, hospitals
   - Entertainment venues
   - POI density

## 🐳 Docker Services

| Service | Description | Port |
|---------|-------------|------|
| `postgres` | PostgreSQL database | 5432 |
| `tor` | Tor proxy for anonymity | 9050, 9051 |
| `server` | API server | 3000 |
| `scraper` | Data scraper (one-time) | - |
| `scheduler` | Periodic scraper | - |

## 🔧 Adding Custom Parsers

1. Create new file in `parsers/`
2. Extend `BaseParser`
3. Implement `Parser` interface
4. Register in `services/scraper.go`

```go
type CustomParser struct {
    *BaseParser
}

func (cp *CustomParser) Name() string {
    return "custom"
}

func (cp *CustomParser) Parse(ctx context.Context) ([]models.Property, error) {
    // Use cp.Fetch() - includes Tor, retry, rate limiting
    body, err := cp.Fetch(ctx, url)
    // Parse and return properties
}
```

📖 **[Parser Development Guide](COMPREHENSIVE_GUIDE.md#creating-custom-parsers)**

## 🚨 Troubleshooting

### Tor not working?
```bash
docker logs pricemap-tor
docker-compose restart tor
```

### Getting blocked?
```bash
# Increase delays
RATE_LIMIT_DELAY=10
MAX_RETRIES=5
```

### No data after scraping?
```bash
docker-compose logs scraper
docker exec -it pricemap-db psql -U postgres -d pricemap -c "SELECT COUNT(*) FROM properties;"
```

📖 **[Complete Troubleshooting Guide](COMPREHENSIVE_GUIDE.md#troubleshooting)**

## 📈 Performance Tuning

**Fast (risky):**
```bash
RATE_LIMIT_DELAY=1
MAX_RETRIES=2
```

**Balanced (recommended):**
```bash
RATE_LIMIT_DELAY=3
MAX_RETRIES=3
```

**Safe (slow):**
```bash
RATE_LIMIT_DELAY=10
MAX_RETRIES=5
```

📖 **[Performance Optimization Guide](COMPREHENSIVE_GUIDE.md#performance-tuning)**

## 📖 Documentation

- **[Comprehensive Guide](COMPREHENSIVE_GUIDE.md)** - Complete documentation
- **[Data Sources](docs/DATA_SOURCES.md)** - Available data sources
- **[Global Coverage](docs/GLOBAL_COVERAGE.md)** - Supported cities
- **[Parser Implementation](docs/PARSER_IMPLEMENTATION.md)** - Creating parsers
- **[Implemented Parsers](docs/IMPLEMENTED_PARSERS.md)** - Current parsers

## 🤝 Contributing

Pull requests welcome! See [Contributing Guide](COMPREHENSIVE_GUIDE.md#contributing).

**Guidelines:**
- Write tests for new features
- Follow Go conventions
- Update documentation
- Keep PRs focused

## 📝 License

MIT License - see [LICENSE](LICENSE) file

## 🙏 Acknowledgments

- OpenStreetMap for free maps
- Nominatim for free geocoding
- Tor Project for anonymity
- Go Community for excellent libraries

---

**Built with ❤️ using Go** | [Report Issues](https://github.com/VadimToptunov/pricemap-go/issues) | [Discussions](https://github.com/VadimToptunov/pricemap-go/discussions)
