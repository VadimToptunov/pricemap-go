# Global Coverage Implementation

This document describes how the system parses real estate data from all populated areas worldwide.

## Architecture

The system is designed to parse data from **all cities and towns** where real estate is sold or rented. This is achieved through:

1. **Multi-city parsing** - Each parser iterates over multiple cities
2. **Sale and rent support** - All parsers support both sale and rental properties
3. **Country-specific parsers** - Dedicated parsers for each country/market
4. **City discovery** - Automatic discovery of cities using OpenStreetMap and curated lists

## Current Coverage

### Implemented Parsers

#### CianParser (Russia)
- **Cities**: 30+ major Russian cities
- **Types**: Apartments, houses, rooms
- **Deal Types**: Sale and rent
- **Coverage**: ~30 cities × 2 deal types × 3 property types = 180+ search combinations

#### RightmoveParser (UK)
- **Cities**: 25+ major UK cities
- **Types**: All property types
- **Deal Types**: Sale and rent
- **Coverage**: ~25 cities × 2 deal types = 50+ search combinations

### City Lists

The system includes curated lists of major cities:

- **North America**: 8 cities (USA, Canada)
- **Europe**: 10 cities (UK, France, Germany, Spain, Italy, Netherlands, Russia, Turkey)
- **Asia**: 9 cities (Japan, China, India, Thailand, Singapore, South Korea, UAE)
- **Oceania**: 3 cities (Australia, New Zealand)
- **South America**: 4 cities (Brazil, Argentina, Peru)
- **Africa**: 4 cities (Egypt, Nigeria, South Africa)

**Total**: 38+ major cities worldwide

## How It Works

### 1. City Iteration

Each parser iterates over a list of cities:

```go
cities := []string{"Moscow", "Saint Petersburg", "Novosibirsk", ...}
for _, city := range cities {
    // Parse properties for this city
}
```

### 2. Deal Type Support

All parsers support both sale and rent:

```go
dealTypes := []string{"sale", "rent"}
for _, dealType := range dealTypes {
    // Parse for this deal type
}
```

### 3. Property Type Coverage

Parsers extract all property types:
- Apartments/flats
- Houses
- Rooms
- Commercial (where applicable)

### 4. Automatic City Discovery

The `utils.CityService` can discover cities from:
- OpenStreetMap Nominatim API
- Curated lists of major cities
- Country-specific city databases

## Adding New Cities

### Method 1: Add to Parser Directly

```go
cities := []string{
    "New City 1",
    "New City 2",
    // ...
}
```

### Method 2: Use City Service

```go
cityService := utils.NewCityService()
cities, err := cityService.GetCitiesFromOpenStreetMap("US", 100000) // Cities with 100k+ population
```

### Method 3: Use Curated Lists

```go
cities := utils.GetMajorCities()
// Filter by country if needed
ukCities := utils.GetCitiesByCountry("United Kingdom")
```

## Expanding Coverage

### To Add a New Country

1. **Create a new parser** in `parsers/` directory
2. **Implement city list** for that country
3. **Add sale and rent support**
4. **Register parser** in `services/scraper.go`

Example structure:

```go
func (p *NewCountryParser) Parse(ctx context.Context) ([]models.Property, error) {
    cities := []string{"City1", "City2", ...}
    dealTypes := []string{"sale", "rent"}
    
    for _, city := range cities {
        for _, dealType := range dealTypes {
            // Parse properties
        }
    }
}
```

### To Add More Cities to Existing Parser

Simply extend the cities list in the parser:

```go
cities := []string{
    // Existing cities
    "New City 1",
    "New City 2",
    // ...
}
```

## Performance Considerations

### Rate Limiting

- **Between cities**: 1-2 seconds delay
- **Between deal types**: 1-2 seconds delay
- **Between property types**: 1-2 seconds delay

### Parallel Processing

Parsers can be run in parallel for different countries:

```go
// Run country parsers in parallel
go cianParser.Parse(ctx)
go rightmoveParser.Parse(ctx)
go zillowParser.Parse(ctx)
```

### Caching

- City coordinates are cached
- Geocoding results are cached (can be implemented)
- Property data is deduplicated by ExternalID

## Data Quality

### Validation

All properties are validated:
- Price > 0
- Valid coordinates (or geocoded address)
- Required fields present

### Deduplication

Properties are deduplicated by:
- Source + ExternalID (unique per source)
- Address + coordinates (fuzzy matching)

### Geocoding

If coordinates are missing:
1. Try to extract from page data
2. Use geocoding service (OpenCage or Nominatim)
3. Skip if geocoding fails (but log for review)

## Monitoring

### Metrics to Track

- Properties parsed per city
- Properties parsed per deal type
- Success rate per parser
- Geocoding success rate
- Average properties per city

### Logging

All parsers log:
- Number of properties found per city
- Errors encountered
- Skipped properties (with reasons)

## Future Enhancements

### Automatic City Discovery

1. Query OpenStreetMap for all cities in a country
2. Filter by population threshold
3. Automatically generate city lists

### Dynamic City Lists

1. Store city lists in database
2. Update from external sources
3. Prioritize cities by activity

### Coverage Expansion

1. Add parsers for more countries
2. Support smaller towns (not just major cities)
3. Add commercial property support
4. Add land/plot support

## Example: Full Coverage Workflow

```
1. Start scraper
2. For each country parser:
   a. Load city list
   b. For each city:
      - Parse sale properties
      - Parse rent properties
      - Geocode addresses if needed
      - Save to database
3. Calculate factors for all properties
4. Generate heatmap data
```

## Current Status

- ✅ Russia: 30+ cities, sale & rent
- ✅ UK: 25+ cities, sale & rent
- ⏳ USA: Parser needed (Zillow, Realtor.com)
- ⏳ Spain: Parser needed (Idealista)
- ⏳ Germany: Parser needed (ImmobilienScout24)
- ⏳ France: Parser needed (Leboncoin)
- ⏳ And 100+ more countries...

## Contributing

To add support for a new country:

1. Identify major real estate websites
2. Create parser following existing patterns
3. Add city list for that country
4. Test with a few cities first
5. Submit pull request

See [PARSER_IMPLEMENTATION.md](PARSER_IMPLEMENTATION.md) for detailed guide.

