# PriceMap Go

A system for parsing real estate prices worldwide with map visualization and analysis of factors affecting prices.

## Features

- 🔍 Parsing real estate data from open sources
- 🗺️ Price visualization on an interactive map (heatmap)
- 📊 Analysis of factors affecting prices:
  - Crime and safety
  - Transportation accessibility
  - School and education ratings
  - Infrastructure (shops, parks, hospitals)
- 🌍 Support for multiple cities and countries
- ⚡ Automatic data updates on schedule

## Project Structure

```
pricemap-go/
├── cmd/
│   ├── server/      # API server
│   ├── scraper/     # Data parser
│   └── scheduler/   # Task scheduler
├── api/             # API handlers and routing
├── models/          # Data models
├── parsers/         # Parsers for various sources
├── services/        # Business logic
├── database/        # Database operations
├── config/          # Configuration
└── web/             # Frontend (HTML/CSS/JS)
```

## Installation

### Requirements

- Go 1.21 or higher
- PostgreSQL 12 or higher
- Google Maps API key (for map)

### Installation Steps

1. Clone the repository:
```bash
git clone <repository-url>
cd pricemap-go
```

2. Install dependencies:
```bash
go mod download
```

3. Set up the database:
```bash
# Create PostgreSQL database
createdb pricemap

# Or use Docker
docker run --name pricemap-db -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=pricemap -p 5432:5432 -d postgres
```

4. Configure environment variables:
```bash
cp .env.example .env
# Edit the .env file and specify your settings
```

5. Run migrations (automatically on first run)

## Usage

### Quick Start with Docker

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Manual Setup

#### Running the API Server

```bash
go run cmd/server/main.go
# Or use Makefile
make run
```

The server will be available at `http://localhost:8080`

#### Running the Scraper (one-time)

```bash
go run cmd/scraper/main.go
# Or use Makefile
make scrape
```

#### Running the Scheduler (automatic updates)

```bash
go run cmd/scheduler/main.go
# Or use Makefile
make schedule
```

The scheduler will automatically run parsing on schedule (default: every 6 hours).

### Using Makefile

```bash
make build      # Build all binaries
make test       # Run tests
make docker-up  # Start Docker containers
make clean      # Clean build artifacts
```

## API Endpoints

### GET /api/v1/heatmap
Get heatmap data

Parameters:
- `lat_min`, `lat_max`, `lng_min`, `lng_max` - area boundaries

### GET /api/v1/properties
Get list of real estate properties

Parameters:
- `city` - filter by city
- `country` - filter by country
- `type` - property type (apartment, house)
- `price_min`, `price_max` - price range
- `page`, `limit` - pagination

### GET /api/v1/properties/:id
Get detailed information about a property

### GET /api/v1/stats
Get statistics

### GET /api/v1/metrics
Get system metrics (parsing stats, performance, etc.)

### GET /api/v1/metrics/parser/:parser
Get metrics for a specific parser

## Data Sources

For a comprehensive list of available data sources for parsing real estate data, see [DATA_SOURCES.md](docs/DATA_SOURCES.md).

The document includes:
- Real estate listing websites by country
- Government open data portals
- Crime data sources
- Transportation data (GTFS, transit APIs)
- Education/school rating sources
- Infrastructure and POI data
- Economic and demographic data

## Global Coverage

The system is designed to parse real estate data from **all populated areas worldwide** where properties are sold or rented.

### Current Coverage

- **Russia**: 30+ cities (Cian.ru) - Sale & Rent
- **United Kingdom**: 25+ cities (Rightmove.co.uk) - Sale & Rent
- **38+ major cities** across 6 continents in curated lists

### Features

- ✅ **Multi-city parsing** - Each parser covers multiple cities
- ✅ **Sale & Rent support** - All parsers extract both sale and rental properties
- ✅ **Automatic city discovery** - Uses OpenStreetMap and curated city lists
- ✅ **Global expansion ready** - Easy to add new countries and cities

See [GLOBAL_COVERAGE.md](docs/GLOBAL_COVERAGE.md) for detailed coverage information.

## Implemented Parsers

The following parsers are currently implemented:

- **CianParser** - cian.ru (Russia) - 30+ cities, sale & rent
- **RightmoveParser** - rightmove.co.uk (UK) - 25+ cities, sale & rent
- **ZillowParser** - zillow.com (USA) - 30+ cities, sale & rent
- **IdealistaParser** - idealista.com (Spain) - 20+ cities, sale & rent
- **OpenDataParser** - Generic parser for government open data portals
- **UniversalParser** - Meta-parser combining multiple sources
- **ExampleParser** - Template for creating new parsers

**Total Coverage**: 105+ cities across 4 countries, 200+ search combinations

See [IMPLEMENTED_PARSERS.md](docs/IMPLEMENTED_PARSERS.md) for details.

## Adding New Parsers

1. Create a new file in `parsers/` (e.g., `parsers/avito.go`)
2. Implement the `Parser` interface:
```go
type Parser interface {
    Name() string
    Parse(ctx context.Context) ([]models.Property, error)
    GetBaseURL() string
}
```
3. Add the parser to `services/scraper.go`:
```go
parsers: []parsers.Parser{
    parsers.NewExampleParser(),
    parsers.NewAvitoParser(), // your new parser
},
```

See [DATA_SOURCES.md](docs/DATA_SOURCES.md) for a list of available sources to implement parsers for.

For implementation guidance, see [PARSER_IMPLEMENTATION.md](docs/PARSER_IMPLEMENTATION.md).

## Price Influencing Factors

The system analyzes the following factors:

1. **Crime** (0-100, where 100 is the safest)
   - Integration with crime data
   - Analysis of area statistics

2. **Transportation Accessibility** (0-100)
   - Proximity to metro/stops
   - Public transport availability
   - Time to city center

3. **Education** (0-100)
   - School ratings in the area
   - Proximity to educational institutions

4. **Infrastructure** (0-100)
   - Shops, parks, hospitals
   - Entertainment venues

## Frontend

Open `web/index.html` in a browser. Make sure:
1. API server is running
2. Correct Google Maps API key is specified in `web/index.html`

## Development

### Adding New Data Sources

The system is designed for easy addition of new parsers. Each parser should:
- Implement the `Parser` interface
- Return structured data in `models.Property` format
- Handle errors correctly

### Integration with External APIs

To improve factor analysis, you can integrate:
- **Crime**: CrimeData.com, local police APIs
- **Transportation**: Google Maps API, OpenStreetMap, GTFS
- **Education**: GreatSchools API, local education data
- **Infrastructure**: Google Places API, Foursquare API

## License

MIT

## Contributing

Pull requests and issues are welcome!
