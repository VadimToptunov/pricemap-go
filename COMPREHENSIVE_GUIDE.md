# PriceMap-Go: Complete Documentation

**Version:** 2.0  
**Last Updated:** December 12, 2025  
**Production Ready:** ✅

A sophisticated real estate price scraping and analysis system with global coverage, Tor integration, and advanced anti-blocking mechanisms.

---

## Table of Contents

1. [Overview](#overview)
2. [Features](#features)
3. [Architecture](#architecture)
4. [Installation](#installation)
5. [Configuration](#configuration)
6. [Usage](#usage)
7. [Parsers](#parsers)
8. [Anti-Blocking Mechanisms](#anti-blocking-mechanisms)
9. [API Reference](#api-reference)
10. [Development](#development)
11. [Troubleshooting](#troubleshooting)
12. [Performance Tuning](#performance-tuning)
13. [Security](#security)
14. [Contributing](#contributing)

---

## Overview

PriceMap-Go is a comprehensive real estate data collection and analysis platform that aggregates property listings from multiple sources worldwide. It features advanced web scraping capabilities with built-in anti-blocking mechanisms, Tor integration, and a sophisticated proxy pool system.

### Key Capabilities

- **Global Coverage**: Parse real estate data from 100+ cities across 4 continents
- **Multiple Sources**: Commercial sites (Cian, Rightmove, Zillow, Idealista) + Open Data portals
- **Anti-Blocking**: Tor integration, proxy pool, user-agent rotation, rate limiting
- **Robust Scraping**: Exponential backoff retry logic, circuit rotation, error handling
- **Real-Time Analysis**: Crime data, transport accessibility, education ratings, infrastructure scores
- **Interactive Visualization**: Heatmaps, property markers, statistical analysis
- **RESTful API**: JSON endpoints for all data with filtering and pagination
- **No API Keys Required**: Uses free OpenStreetMap and Nominatim services

---

## Features

### Core Features

#### 🌍 Global Real Estate Scraping
- **30+ cities in Russia** via Cian.ru
- **25+ cities in UK** via Rightmove.co.uk
- **30+ cities in USA** via Zillow.com
- **20+ cities in Spain** via Idealista.com
- **7 major cities** via Open Data portals (NYC, London, Berlin, Paris, Tokyo, Sydney, Moscow)

#### 🔒 Advanced Anti-Blocking

**Tor Integration:**
- Automatic circuit rotation every 10 requests
- Manual circuit rotation on errors
- SOCKS5 proxy support
- Control port integration for IP rotation

**Proxy Pool:**
- Support for HTTP/HTTPS/SOCKS5 proxies
- Round-robin and random selection
- Automatic health checking
- Failed proxy removal
- Multiple proxy protocol support

**Request Randomization:**
- 10 different browser User-Agent strings
- Random delays between requests (3-5 seconds default)
- Realistic browser headers
- Cookie and session management

**Retry Logic:**
- Exponential backoff (2^n × base delay)
- 3 attempts by default
- Automatic Tor rotation on retry
- HTTP 429 and 5xx handling

#### 📊 Data Analysis

**Price Factors:**
- **Crime Safety Score** (0-100): Integration with police APIs and crime databases
- **Transport Accessibility** (0-100): Distance to metro/transit, GTFS data
- **Education Rating** (0-100): School ratings and proximity
- **Infrastructure Score** (0-100): Shops, parks, hospitals, entertainment

**Currency Support:**
- Automatic conversion to USD
- Support for RUB, GBP, EUR, USD, and more
- Real-time exchange rates

#### 🗺️ Visualization

**Interactive Map (Leaflet/OpenStreetMap):**
- Heatmap layer for price density
- Property markers with clustering
- Filter by price, type, location
- No API keys required

**Statistical Dashboard:**
- Total properties count
- Average, min, max prices
- Price distribution charts
- Factor analysis graphs

---

## Architecture

### System Components

```
┌─────────────────────────────────────────────────────────────┐
│                         Frontend                             │
│  (HTML/CSS/JS + Leaflet + OpenStreetMap)                    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                      API Server (Go)                         │
│  • REST endpoints                                            │
│  • CORS, rate limiting                                       │
│  • Metrics collection                                        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                    Services Layer                            │
│  • ScraperService  • FactorsService                          │
│  • MetricsService  • CacheService                            │
└─────────────────────────────────────────────────────────────┐
                              │
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                     Parsers (with Tor/Proxy)                 │
│  • BaseParser (shared logic)                                 │
│  • CianParser, RightmoveParser, ZillowParser, etc.          │
│  • OpenData parsers (NYC, London, etc.)                      │
└─────────────────────────────────────────────────────────────┐
                              │
                              ↓
┌──────────────┬──────────────┬────────────────┬──────────────┐
│   Tor Proxy  │  Proxy Pool  │  Rate Limiter  │  User Agent  │
│   (Circuit   │  (HTTP/      │  (Random       │  (Rotation)  │
│   Rotation)  │  SOCKS5)     │  Delays)       │              │
└──────────────┴──────────────┴────────────────┴──────────────┘
                              │
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                    Database (PostgreSQL)                     │
│  • Properties  • Factors  • Metrics                          │
└─────────────────────────────────────────────────────────────┘
```

### Directory Structure

```
pricemap-go/
├── cmd/
│   ├── server/         # API server entry point
│   ├── scraper/        # One-time scraping job
│   ├── scheduler/      # Periodic scraping scheduler
│   └── geocode/        # Geocoding utility
├── api/
│   ├── handlers.go     # HTTP request handlers
│   ├── middleware.go   # CORS, rate limiting, logging
│   ├── router.go       # Route definitions
│   └── metrics.go      # Metrics endpoints
├── models/
│   └── property.go     # Data models
├── parsers/
│   ├── base.go         # BaseParser with Tor/retry logic
│   ├── cian.go         # Russia (Cian.ru)
│   ├── rightmove.go    # UK (Rightmove.co.uk)
│   ├── zillow.go       # USA (Zillow.com)
│   ├── idealista.go    # Spain (Idealista.com)
│   └── *_opendata.go   # Open Data parsers
├── services/
│   ├── scraper.go      # Scraping orchestration
│   ├── factors.go      # Factor calculation
│   ├── metrics.go      # Performance tracking
│   └── cache.go        # In-memory caching
├── utils/
│   ├── tor.go          # Tor circuit rotation
│   ├── proxy_pool.go   # Proxy management
│   ├── useragent.go    # User-Agent rotation
│   ├── geocoding.go    # Free geocoding (Nominatim)
│   ├── currency.go     # Currency conversion
│   ├── validation.go   # Data validation
│   └── cities.go       # City lists
├── database/
│   └── database.go     # DB connection & migrations
├── config/
│   └── config.go       # Configuration management
├── web/
│   ├── index.html      # Frontend
│   ├── style.css       # Styles
│   └── app-leaflet.js  # Map logic
├── docs/               # Documentation
├── docker-compose.yml  # Docker orchestration
├── Dockerfile          # Multi-stage build
├── Makefile           # Build commands
├── env.example        # Configuration template
└── go.mod             # Go dependencies
```

---

## Installation

### Prerequisites

- **Go 1.21+** (for local development)
- **Docker & Docker Compose** (recommended)
- **PostgreSQL 12+** (provided via Docker)

### Quick Start (Docker - Recommended)

```bash
# 1. Clone repository
git clone https://github.com/VadimToptunov/pricemap-go.git
cd pricemap-go

# 2. Create configuration
cp env.example .env

# 3. Start all services (including Tor)
docker-compose up -d

# 4. Check status
docker-compose ps

# 5. View logs
docker-compose logs -f scraper
```

### Local Development Setup

```bash
# 1. Install Go dependencies
go mod download

# 2. Start PostgreSQL (Docker)
docker-compose up postgres -d

# 3. Start Tor (optional, for testing)
docker-compose up tor -d

# 4. Create .env file
cp env.example .env
# Edit .env with your local settings

# 5. Run migrations (automatic on first start)
go run cmd/server/main.go

# 6. Run scraper
go run cmd/scraper/main.go

# 7. Run API server
go run cmd/server/main.go
```

### Building from Source

```bash
# Build all binaries
make build

# Run binaries
./bin/server
./bin/scraper
./bin/scheduler
```

---

## Configuration

### Environment Variables

Create a `.env` file from `env.example`:

```bash
# Database Configuration
DB_HOST=localhost              # Database host
DB_PORT=5432                   # PostgreSQL port
DB_USER=postgres               # Database user
DB_PASSWORD=postgres           # Database password
DB_NAME=pricemap               # Database name

# Server Configuration
SERVER_PORT=3000               # API server port

# API Keys (ALL OPTIONAL - not required!)
GOOGLE_MAPS_API_KEY=           # Not used (OpenStreetMap instead)
OPENCAGE_API_KEY=              # Optional (Nominatim fallback)

# Scraping Configuration
USER_AGENT=Mozilla/5.0...      # Default user agent
REQUEST_TIMEOUT=30             # Request timeout in seconds

# Tor Proxy Configuration (ENABLED BY DEFAULT)
USE_TOR=true                   # Enable/disable Tor
TOR_PROXY_HOST=tor             # Tor host (use "localhost" for local)
TOR_PROXY_PORT=9050            # Tor SOCKS5 proxy port
TOR_CONTROL_PORT=9051          # Tor control port
TOR_CONTROL_PASSWORD=          # Leave empty for default setup

# Rate Limiting & Retry
RATE_LIMIT_DELAY=3             # Base delay between requests (seconds)
MAX_RETRIES=3                  # Number of retry attempts
RETRY_DELAY=5                  # Base delay for exponential backoff

# Scheduler
CRON_SCHEDULE=0 */6 * * *      # Cron expression (every 6 hours)
```

### Configuration in Code

The `config.Config` struct in `config/config.go`:

```go
type Config struct {
    // Database
    DBHost, DBPort, DBUser, DBPassword, DBName string
    
    // Server
    ServerPort string
    
    // API Keys (optional)
    GoogleMapsAPIKey, OpenCageAPIKey string
    
    // Scraping
    UserAgent      string
    RequestTimeout int
    
    // Tor
    UseTor            bool
    TorProxyHost      string
    TorProxyPort      string
    TorControlPort    string
    TorControlPassword string
    
    // Rate Limiting
    RateLimitDelay int
    MaxRetries     int
    RetryDelay     int
    
    // Scheduler
    CronSchedule string
}
```

### Docker Compose Configuration

Key services in `docker-compose.yml`:

```yaml
services:
  postgres:      # PostgreSQL database
  tor:           # Tor proxy for anonymity
  server:        # API server
  scraper:       # Data scraper (one-time)
  scheduler:     # Periodic scraper
```

---

## Usage

### Running Services

#### With Docker

```bash
# Start all services
docker-compose up -d

# Start specific service
docker-compose up server -d
docker-compose up scraper -d

# View logs
docker-compose logs -f server
docker-compose logs -f scraper
docker-compose logs -f tor

# Stop services
docker-compose down

# Restart service
docker-compose restart scraper
```

#### Local Development

```bash
# Terminal 1: API Server
go run cmd/server/main.go

# Terminal 2: Scraper (one-time)
go run cmd/scraper/main.go

# Terminal 3: Scheduler (continuous)
go run cmd/scheduler/main.go
```

### Using Makefile

```bash
make build          # Build all binaries
make run            # Run API server
make scrape         # Run scraper once
make schedule       # Run scheduler
make test           # Run tests
make test-coverage  # Run tests with coverage
make docker-up      # Start Docker services
make docker-down    # Stop Docker services
make clean          # Clean build artifacts
```

### Accessing the Application

**API Server:**
```
http://localhost:3000
```

**Frontend:**
Open `web/index.html` in a browser (API server must be running)

**API Endpoints:**
- `GET /api/v1/properties` - List all properties
- `GET /api/v1/properties/:id` - Get property details
- `GET /api/v1/heatmap` - Get heatmap data
- `GET /api/v1/stats` - Get statistics
- `GET /api/v1/metrics` - Get system metrics
- `GET /api/v1/metrics/parser/:parser` - Get parser-specific metrics

---

## Parsers

### Available Parsers

#### Commercial Real Estate Sites

**1. CianParser (Russia)**
- **Source:** cian.ru
- **Coverage:** 30+ major Russian cities
- **Types:** Apartments, houses, rooms
- **Deal Types:** Sale and rent
- **Features:** Price in RUB, area in sq.m, room count

**2. RightmoveParser (UK)**
- **Source:** rightmove.co.uk
- **Coverage:** 25+ UK cities
- **Types:** Properties, apartments
- **Deal Types:** Sale and rent
- **Features:** Price in GBP, area conversion (sq.ft → sq.m)

**3. ZillowParser (USA)**
- **Source:** zillow.com
- **Coverage:** 30+ US cities
- **Types:** Homes, apartments
- **Deal Types:** Sale and rent
- **Features:** Price in USD, bedrooms, bathrooms

**4. IdealistaParser (Spain)**
- **Source:** idealista.com
- **Coverage:** 20+ Spanish cities
- **Types:** Apartments, houses
- **Deal Types:** Sale and rent
- **Features:** Price in EUR, room count (hab/dorm)

#### Open Data Parsers

**5. NYCOpenDataParser**
- **Source:** NYC Open Data API
- **Dataset:** Property sales records
- **Format:** JSON API
- **Coverage:** New York City

**6. LondonOpenDataParser**
- **Source:** London Data Store
- **Dataset:** House prices
- **Format:** JSON
- **Coverage:** Greater London

**7. BerlinOpenDataParser**
- **Source:** Berlin Open Data
- **Dataset:** Real estate market
- **Coverage:** Berlin

**8. ParisOpenDataParser**
- **Source:** Paris Open Data
- **Dataset:** Rent control data
- **Coverage:** Paris

**9. TokyoOpenDataParser**
- **Source:** Tokyo Open Data
- **Dataset:** Property prices
- **Coverage:** Tokyo

**10. SydneyOpenDataParser**
- **Source:** NSW Open Data
- **Dataset:** Property sales
- **Coverage:** Sydney

**11. MoscowOpenDataParser**
- **Source:** Moscow Open Data (data.mos.ru)
- **Dataset:** Real estate transactions
- **Coverage:** Moscow

### Creating Custom Parsers

#### Step 1: Implement Parser Interface

```go
package parsers

import (
    "context"
    "pricemap-go/models"
)

type CustomParser struct {
    *BaseParser
    geocoding *utils.GeocodingService
}

func NewCustomParser() *CustomParser {
    return &CustomParser{
        BaseParser: NewBaseParser("https://example.com"),
        geocoding:  utils.NewGeocodingService(),
    }
}

func (cp *CustomParser) Name() string {
    return "custom"
}

func (cp *CustomParser) Parse(ctx context.Context) ([]models.Property, error) {
    var properties []models.Property
    
    // 1. Build URL
    url := cp.baseURL + "/listings"
    
    // 2. Fetch (includes Tor, retry, rate limiting)
    body, err := cp.Fetch(ctx, url)
    if err != nil {
        return nil, err
    }
    defer body.Close()
    
    // 3. Parse HTML/JSON
    doc, err := goquery.NewDocumentFromReader(body)
    if err != nil {
        return nil, err
    }
    
    // 4. Extract properties
    doc.Find(".listing").Each(func(i int, s *goquery.Selection) {
        property := cp.parseProperty(s)
        if property != nil {
            properties = append(properties, *property)
        }
    })
    
    return properties, nil
}

func (cp *CustomParser) parseProperty(s *goquery.Selection) *models.Property {
    // Extract data from HTML
    return &models.Property{
        Source:     cp.Name(),
        ExternalID: "...",
        Price:      parsePrice(s.Find(".price").Text()),
        Address:    s.Find(".address").Text(),
        // ... more fields
        ScrapedAt:  time.Now(),
        IsActive:   true,
    }
}
```

#### Step 2: Register Parser

Add to `services/scraper.go`:

```go
func NewScraperService() *ScraperService {
    return &ScraperService{
        parsers: []parsers.Parser{
            parsers.NewCianParser(),
            parsers.NewCustomParser(), // Add here
            // ...
        },
        // ...
    }
}
```

### Parser Best Practices

1. **Always use BaseParser.Fetch()** - includes retry, rate limiting, Tor
2. **Check context cancellation** - support graceful shutdown
3. **Validate extracted data** - skip invalid properties
4. **Use geocoding sparingly** - rate limits apply
5. **Log progress** - help debugging
6. **Handle errors gracefully** - don't fail entire batch

---

## Anti-Blocking Mechanisms

### Tor Integration

**How it works:**
1. All requests route through Tor SOCKS5 proxy (port 9050)
2. Circuit rotates automatically every 10 requests
3. Manual rotation on HTTP errors (429, 5xx)
4. Control port (9051) used for NEWNYM signal

**Configuration:**

```bash
USE_TOR=true
TOR_PROXY_HOST=tor              # Docker service name
TOR_PROXY_PORT=9050             # SOCKS5 port
TOR_CONTROL_PORT=9051           # Control port
TOR_CONTROL_PASSWORD=           # Empty for Docker setup
```

**Testing Tor:**

```bash
# Check Tor is running
docker logs pricemap-tor

# Test IP rotation
curl --socks5 localhost:9050 https://api.ipify.org
# Run again to see IP change
```

**Tor Rotation Code:**

```go
// utils/tor.go
func (tc *TorController) RotateCircuit() error {
    // 1. Connect to control port
    conn, err := net.Dial("tcp", tc.controlAddr)
    
    // 2. Authenticate
    fmt.Fprintf(conn, "AUTHENTICATE\r\n")
    
    // 3. Send NEWNYM signal
    fmt.Fprintf(conn, "SIGNAL NEWNYM\r\n")
    
    // 4. Wait for new circuit
    time.Sleep(2 * time.Second)
    
    return nil
}
```

### Proxy Pool

**Features:**
- Support for HTTP, HTTPS, SOCKS5 proxies
- Round-robin and random selection
- Automatic health checking
- Failed proxy removal
- Per-proxy failure tracking

**Usage:**

```go
// Create proxy pool
pool := utils.NewProxyPool()

// Add proxies
pool.AddProxy("socks5://proxy1.com:1080", "socks5")
pool.AddProxy("http://proxy2.com:8080", "http")

// Get next proxy (round-robin)
proxy, err := pool.GetNextProxy()

// Create HTTP client with proxy
client, err := pool.CreateHTTPClient(proxy, 30*time.Second)

// Mark proxy status
pool.MarkProxyWorking(proxy)    // Success
pool.MarkProxyFailed(proxy)     // Failed

// Health check all proxies
pool.CheckProxies("https://example.com")

// Remove dead proxies
pool.RemoveFailedProxies()
```

### User-Agent Rotation

**Built-in User-Agents (10 total):**
- Chrome on macOS, Windows, Linux
- Firefox on macOS, Windows, Linux
- Safari on macOS
- Edge on Windows

**Usage:**

```go
import "pricemap-go/utils"

// Get random User-Agent
ua := utils.GetRandomUserAgent()

// Set in request
req.Header.Set("User-Agent", ua)
```

### Rate Limiting

**Dynamic Delays:**
```go
// Base delay + random (0-2 sec)
delay := utils.GetRandomDelay(
    config.AppConfig.RateLimitDelay,
    config.AppConfig.RateLimitDelay + 2,
)
time.Sleep(delay)
```

**Configuration:**
```bash
RATE_LIMIT_DELAY=3    # 3-5 seconds between requests
```

### Retry Logic with Exponential Backoff

**Algorithm:**
```
Attempt 0: immediate
Attempt 1: wait 5 seconds  (2^0 × 5)
Attempt 2: wait 10 seconds (2^1 × 5)
Attempt 3: wait 20 seconds (2^2 × 5)
```

**Configuration:**
```bash
MAX_RETRIES=3      # Number of retries
RETRY_DELAY=5      # Base delay (seconds)
```

**Code:**
```go
for attempt := 0; attempt <= maxRetries; attempt++ {
    body, err := bp.fetchWithRetry(ctx, url, attempt)
    if err == nil {
        return body, nil
    }
    
    if attempt < maxRetries {
        backoff := time.Duration(math.Pow(2, float64(attempt))) * 
                   time.Second * time.Duration(config.AppConfig.RetryDelay)
        time.Sleep(backoff)
        
        // Rotate Tor on retry
        if bp.torController != nil {
            bp.torController.RotateCircuit()
        }
    }
}
```

### Request Headers

**Realistic Browser Headers:**
```go
req.Header.Set("User-Agent", randomUA)
req.Header.Set("Accept", "text/html,application/xhtml+xml,...")
req.Header.Set("Accept-Language", "en-US,en;q=0.9,ru;q=0.8")
req.Header.Set("Connection", "keep-alive")
req.Header.Set("Upgrade-Insecure-Requests", "1")
req.Header.Set("Sec-Fetch-Dest", "document")
req.Header.Set("Sec-Fetch-Mode", "navigate")
req.Header.Set("Sec-Fetch-Site", "none")
req.Header.Set("Cache-Control", "max-age=0")
req.Header.Set("DNT", "1")
```

---

## API Reference

### Base URL
```
http://localhost:3000/api/v1
```

### Endpoints

#### 1. List Properties

**GET** `/properties`

**Query Parameters:**
- `city` (string) - Filter by city name
- `country` (string) - Filter by country
- `type` (string) - Filter by type (apartment, house, room)
- `price_min` (float) - Minimum price
- `price_max` (float) - Maximum price
- `page` (int) - Page number (default: 1)
- `limit` (int) - Items per page (default: 50, max: 100)

**Example:**
```bash
curl "http://localhost:3000/api/v1/properties?city=Moscow&type=apartment&limit=10"
```

**Response:**
```json
{
  "properties": [
    {
      "id": 1,
      "source": "cian",
      "external_id": "123456",
      "title": "2-room apartment",
      "price": 5000000,
      "currency": "RUB",
      "address": "Tverskaya St, 15",
      "city": "Moscow",
      "country": "Russia",
      "latitude": 55.7558,
      "longitude": 37.6173,
      "area": 65.5,
      "rooms": 2,
      "bedrooms": 2,
      "type": "apartment",
      "url": "https://cian.ru/...",
      "scraped_at": "2025-12-12T10:30:00Z",
      "is_active": true
    }
  ],
  "total": 150,
  "page": 1,
  "limit": 10,
  "pages": 15
}
```

#### 2. Get Property Details

**GET** `/properties/:id`

**Example:**
```bash
curl "http://localhost:3000/api/v1/properties/1"
```

**Response:**
```json
{
  "id": 1,
  "source": "cian",
  "external_id": "123456",
  "price": 5000000,
  "currency": "RUB",
  "address": "Tverskaya St, 15",
  "city": "Moscow",
  "country": "Russia",
  "latitude": 55.7558,
  "longitude": 37.6173,
  "area": 65.5,
  "rooms": 2,
  "factors": {
    "crime_score": 85,
    "transport_score": 90,
    "education_score": 88,
    "infrastructure_score": 92
  }
}
```

#### 3. Get Heatmap Data

**GET** `/heatmap`

**Query Parameters:**
- `lat_min` (float) - Minimum latitude
- `lat_max` (float) - Maximum latitude
- `lng_min` (float) - Minimum longitude
- `lng_max` (float) - Maximum longitude

**Example:**
```bash
curl "http://localhost:3000/api/v1/heatmap?lat_min=55.7&lat_max=55.8&lng_min=37.5&lng_max=37.7"
```

**Response:**
```json
{
  "points": [
    {
      "latitude": 55.7558,
      "longitude": 37.6173,
      "intensity": 5000000
    }
  ]
}
```

#### 4. Get Statistics

**GET** `/stats`

**Example:**
```bash
curl "http://localhost:3000/api/v1/stats"
```

**Response:**
```json
{
  "total_properties": 15420,
  "average_price": 3500000,
  "min_price": 500000,
  "max_price": 50000000,
  "cities_count": 85,
  "countries_count": 4,
  "sources": {
    "cian": 8500,
    "rightmove": 3200,
    "zillow": 2500,
    "idealista": 1220
  },
  "last_update": "2025-12-12T10:30:00Z"
}
```

#### 5. Get System Metrics

**GET** `/metrics`

**Example:**
```bash
curl "http://localhost:3000/api/v1/metrics"
```

**Response:**
```json
{
  "parsers": [
    {
      "name": "cian",
      "total_runs": 50,
      "total_properties": 8500,
      "total_saved": 8200,
      "total_errors": 150,
      "average_duration": "45m30s",
      "last_run": "2025-12-12T10:00:00Z",
      "success_rate": 96.5
    }
  ],
  "system": {
    "uptime": "72h15m30s",
    "goroutines": 12,
    "memory_mb": 245
  }
}
```

#### 6. Get Parser-Specific Metrics

**GET** `/metrics/parser/:parser`

**Example:**
```bash
curl "http://localhost:3000/api/v1/metrics/parser/cian"
```

---

## Development

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./utils/...

# With verbose output
go test -v ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Using Makefile
make test
make test-coverage
```

### Adding Dependencies

```bash
# Add dependency
go get github.com/example/package

# Update dependencies
go mod tidy

# Vendor dependencies (optional)
go mod vendor
```

### Code Structure Guidelines

**Models** (`models/`):
- Define data structures
- Include JSON/DB tags
- Add validation methods

**Parsers** (`parsers/`):
- Extend BaseParser
- Implement Parser interface
- Use Fetch() for HTTP requests
- Handle context cancellation

**Services** (`services/`):
- Business logic
- Orchestration
- Data processing

**Utils** (`utils/`):
- Helper functions
- Reusable components
- No business logic

**API** (`api/`):
- HTTP handlers
- Request/response formatting
- Middleware

### Database Migrations

Migrations run automatically on first server start. Schema:

```sql
-- Properties table
CREATE TABLE properties (
    id SERIAL PRIMARY KEY,
    source VARCHAR(50) NOT NULL,
    external_id VARCHAR(255) NOT NULL,
    title TEXT,
    price DECIMAL(15,2),
    currency VARCHAR(3),
    address TEXT,
    city VARCHAR(100),
    country VARCHAR(100),
    district VARCHAR(100),
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    area DECIMAL(10,2),
    rooms INT,
    bedrooms INT,
    bathrooms INT,
    type VARCHAR(50),
    url TEXT,
    scraped_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE,
    UNIQUE(source, external_id)
);

-- Factors table
CREATE TABLE factors (
    id SERIAL PRIMARY KEY,
    property_id INT REFERENCES properties(id),
    crime_score INT,
    transport_score INT,
    education_score INT,
    infrastructure_score INT,
    calculated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_properties_city ON properties(city);
CREATE INDEX idx_properties_price ON properties(price);
CREATE INDEX idx_properties_location ON properties(latitude, longitude);
CREATE INDEX idx_factors_property ON factors(property_id);
```

---

## Troubleshooting

### Common Issues

#### 1. Tor Connection Failed

**Error:** `Failed to connect to Tor proxy`

**Solutions:**
```bash
# Check Tor is running
docker ps | grep tor

# View Tor logs
docker logs pricemap-tor

# Restart Tor
docker-compose restart tor

# Test Tor manually
curl --socks5 localhost:9050 https://api.ipify.org
```

#### 2. Rate Limiting / Blocking

**Error:** `HTTP 429 Too Many Requests`

**Solutions:**
- Increase `RATE_LIMIT_DELAY` (e.g., to 5 or 10 seconds)
- Verify Tor is enabled and rotating
- Add more proxies to proxy pool
- Reduce number of concurrent scrapers

```bash
# In .env
RATE_LIMIT_DELAY=10
MAX_RETRIES=5
```

#### 3. Database Connection Failed

**Error:** `Failed to connect to database`

**Solutions:**
```bash
# Check PostgreSQL is running
docker ps | grep postgres

# View DB logs
docker logs pricemap-db

# Test connection
docker exec -it pricemap-db psql -U postgres -d pricemap

# Reset database
docker-compose down -v
docker-compose up postgres -d
```

#### 4. Geocoding Errors

**Error:** `Failed to geocode address`

**Solutions:**
- Nominatim has rate limits (1 req/sec)
- Add delay between geocoding requests
- Consider using OpenCage API (optional)
- Many properties will have coordinates from source

```bash
# Optional: Add OpenCage API key
OPENCAGE_API_KEY=your_key_here
```

#### 5. No Data After Scraping

**Issue:** Scraper runs but no properties in DB

**Solutions:**
```bash
# Check scraper logs
docker-compose logs scraper

# Verify parsers are registered
grep "NewCianParser" services/scraper.go

# Test parser individually
go run cmd/scraper/main.go 2>&1 | grep "Found.*properties"

# Check database
docker exec -it pricemap-db psql -U postgres -d pricemap -c "SELECT COUNT(*) FROM properties;"
```

### Debug Mode

Enable verbose logging:

```go
// In main.go
log.SetFlags(log.LstdFlags | log.Lshortfile)
log.SetOutput(os.Stdout)
```

### Performance Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof ./...
go tool pprof mem.prof
```

---

## Performance Tuning

### Scraping Performance

**Configuration:**

```bash
# Faster (risky - may get blocked)
RATE_LIMIT_DELAY=1
MAX_RETRIES=2
RETRY_DELAY=3

# Balanced (recommended)
RATE_LIMIT_DELAY=3
MAX_RETRIES=3
RETRY_DELAY=5

# Slower (very safe)
RATE_LIMIT_DELAY=10
MAX_RETRIES=5
RETRY_DELAY=10
```

**Parser Optimization:**

Reduce cities in parser:
```go
// parsers/cian.go
cities := []string{
    "Moscow", "Saint Petersburg", "Novosibirsk", // Top 3 only
}
```

### Database Performance

**Indexes:**
```sql
CREATE INDEX CONCURRENTLY idx_properties_city ON properties(city);
CREATE INDEX CONCURRENTLY idx_properties_price ON properties(price);
CREATE INDEX CONCURRENTLY idx_properties_location ON properties USING GIST (point(longitude, latitude));
```

**Connection Pool:**
```go
// database/database.go
sqlDB.SetMaxOpenConns(25)
sqlDB.SetMaxIdleConns(10)
sqlDB.SetConnMaxLifetime(5 * time.Minute)
```

### Caching

**In-memory cache** (already implemented):
```go
// services/cache.go
cacheService := NewCacheService(1 * time.Hour) // TTL
```

**Redis** (future enhancement):
```yaml
# docker-compose.yml
redis:
  image: redis:7-alpine
  ports:
    - "6379:6379"
```

### Concurrent Scraping

Run multiple scrapers:
```bash
# Terminal 1
docker-compose up scraper

# Terminal 2 (different parser)
USE_PARSER=rightmove docker-compose up scraper
```

### API Performance

**Pagination:**
```bash
# Use smaller page sizes
curl "localhost:3000/api/v1/properties?limit=20"
```

**Filtering:**
```bash
# Filter at API level, not client
curl "localhost:3000/api/v1/properties?city=Moscow&price_max=10000000"
```

---

## Security

### Best Practices

1. **Never commit `.env` files**
   - Use `env.example` as template
   - `.env` is in `.gitignore`

2. **Use Tor for anonymity**
   - Enabled by default
   - Automatic circuit rotation

3. **Database credentials**
   - Change default passwords in production
   - Use strong passwords
   - Limit database network exposure

4. **API Security**
   - Add authentication (JWT recommended)
   - Enable rate limiting
   - Use HTTPS in production

5. **Docker security**
   - Don't run as root
   - Use non-privileged ports
   - Keep images updated

### Production Checklist

- [ ] Change database password
- [ ] Enable HTTPS (use reverse proxy like Nginx)
- [ ] Add API authentication
- [ ] Configure firewall
- [ ] Set up monitoring (Prometheus/Grafana)
- [ ] Configure backups
- [ ] Set up logging (ELK stack)
- [ ] Use production Tor configuration
- [ ] Add multiple proxy sources
- [ ] Configure domain and DNS
- [ ] Set up SSL certificates (Let's Encrypt)

---

## Contributing

### How to Contribute

1. **Fork the repository**
2. **Create feature branch:** `git checkout -b feature/amazing-feature`
3. **Commit changes:** `git commit -m 'Add amazing feature'`
4. **Push to branch:** `git push origin feature/amazing-feature`
5. **Open Pull Request**

### Contribution Guidelines

- Write tests for new features
- Follow Go conventions (`gofmt`, `golint`)
- Update documentation
- Add comments for complex logic
- Keep PRs focused and small

### Adding New Parsers

See [Creating Custom Parsers](#creating-custom-parsers) section above.

**Checklist:**
- [ ] Implement Parser interface
- [ ] Use BaseParser for HTTP requests
- [ ] Add tests
- [ ] Register in ScraperService
- [ ] Update documentation
- [ ] Add example in README

### Testing Your Changes

```bash
# Run all tests
make test

# Run specific tests
go test ./parsers/... -v

# Check code formatting
gofmt -s -w .

# Run linter
golangci-lint run
```

---

## License

MIT License - see [LICENSE](LICENSE) file for details.

---

## Support

**Issues:** https://github.com/VadimToptunov/pricemap-go/issues  
**Discussions:** https://github.com/VadimToptunov/pricemap-go/discussions  

---

## Acknowledgments

- **OpenStreetMap** for free map tiles
- **Nominatim** for free geocoding
- **Tor Project** for anonymity tools
- **Go Community** for excellent libraries

---

**Built with ❤️ using Go**

