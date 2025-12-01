# Implemented Parsers

This document lists all currently implemented parsers and their status.

## Active Parsers

### 1. CianParser (`parsers/cian.go`)
- **Source**: cian.ru (Russia)
- **Status**: ✅ Implemented
- **Coverage**: Russia-wide
- **Data Extracted**:
  - Price (RUB)
  - Address and location
  - Property type (apartment, house, room)
  - Area (m²)
  - Number of rooms
  - Coordinates (with geocoding fallback)
- **Notes**: 
  - Handles Russian text and currency
  - Includes geocoding for addresses without coordinates
  - Respects rate limiting

### 2. RightmoveParser (`parsers/rightmove.go`)
- **Source**: rightmove.co.uk (UK)
- **Status**: ✅ Implemented
- **Coverage**: UK-wide
- **Data Extracted**:
  - Price (GBP)
  - Address and location
  - Property type
  - Area (sq ft / sq m)
  - Number of bedrooms
  - Coordinates (with geocoding fallback)
- **Notes**:
  - Handles UK property formats
  - Converts sq ft to sq m automatically

### 3. OpenDataParser (`parsers/opendata.go`)
- **Source**: Government open data portals (generic)
- **Status**: ✅ Implemented (JSON support)
- **Coverage**: Configurable per portal
- **Data Extracted**:
  - Price
  - Address
  - Coordinates
  - Area
  - Rooms
  - Property type
- **Notes**:
  - Generic parser for any JSON-based open data API
  - Can be configured for different countries/cities
  - CSV support pending

### 4. ExampleParser (`parsers/example_parser.go`)
- **Source**: Example template
- **Status**: ✅ Template implementation
- **Coverage**: N/A
- **Notes**: Use as a template for new parsers

## Supporting Services

### GeocodingService (`utils/geocoding.go`)
- **Status**: ✅ Implemented
- **Features**:
  - OpenCage API support (with API key)
  - Nominatim (OpenStreetMap) fallback (free, rate-limited)
  - Automatic address to coordinates conversion
- **Usage**: Automatically used by parsers when coordinates are missing

### TransportService (`services/transport.go`)
- **Status**: ✅ Implemented
- **Features**:
  - GTFS data parsing
  - Distance calculation (Haversine formula)
  - Transportation score calculation
  - Support for metro, bus, tram stops
- **Notes**: Ready for integration with GTFS feeds

## Usage

### Adding a Parser

1. Create a new file in `parsers/` directory
2. Implement the `Parser` interface:
   ```go
   type Parser interface {
       Name() string
       Parse(ctx context.Context) ([]models.Property, error)
       GetBaseURL() string
   }
   ```
3. Register in `services/scraper.go`:
   ```go
   parsers: []parsers.Parser{
       parsers.NewYourParser(),
   },
   ```

### Example: Using OpenDataParser

```go
// For a specific city's open data portal
parser := parsers.NewOpenDataParser(
    "https://data.cityofnewyork.us/api/properties",
    "USA",
    "New York",
)
```

## Testing Parsers

To test a parser:

```bash
# Run scraper with specific parser
go run cmd/scraper/main.go
```

Or create unit tests in `parsers/*_test.go`.

## Next Steps

### High Priority
- [ ] Add more country-specific parsers (Zillow for USA, Idealista for Spain, etc.)
- [ ] Implement CSV parsing in OpenDataParser
- [ ] Add error handling and retry logic
- [ ] Add rate limiting per parser

### Medium Priority
- [ ] Add caching for geocoding requests
- [ ] Implement parser health checks
- [ ] Add metrics and monitoring
- [ ] Create parser configuration system

### Low Priority
- [ ] Add support for XML data sources
- [ ] Implement parser discovery system
- [ ] Add parser performance benchmarks

## Parser Best Practices

1. **Rate Limiting**: Always add delays between requests
2. **Error Handling**: Don't fail entire scraping on single property errors
3. **Data Validation**: Validate all extracted data before saving
4. **Geocoding**: Use geocoding service for addresses without coordinates
5. **Logging**: Log important events (parsed count, errors, etc.)
6. **Respect ToS**: Check and respect Terms of Service
7. **User-Agent**: Always set proper User-Agent header

## Known Issues

- Cian parser selectors may need adjustment based on site updates
- Rightmove parser needs testing with actual site structure
- OpenDataParser CSV support not yet implemented
- Geocoding rate limits may affect performance (consider caching)

