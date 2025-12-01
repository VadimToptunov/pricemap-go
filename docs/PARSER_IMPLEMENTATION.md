# Parser Implementation Guide

This guide helps you implement parsers for various real estate data sources.

## Quick Reference: Top Priority Sources

### Real Estate Listings

| Source | Country | Type | Difficulty | Coverage |
|--------|---------|------|------------|----------|
| Zillow | USA | Web scraping | Medium | Excellent |
| Rightmove | UK | Web scraping | Medium | Excellent |
| Циан (Cian) | Russia | Web scraping | Medium | Excellent |
| Realtor.com | USA | Web scraping | Medium | Excellent |
| Idealista | Spain | Web scraping | Medium | Excellent |
| ImmobilienScout24 | Germany | Web scraping | Medium | Excellent |

### Open Data (Easiest to Start)

| Source | Type | Difficulty | Coverage |
|--------|------|------------|----------|
| Data.gov (USA) | API/CSV | Easy | USA cities |
| Data.gov.uk | API/CSV | Easy | UK-wide |
| OpenStreetMap | API | Easy | Worldwide |
| GTFS Feeds | CSV/JSON | Easy | Cities with transit |

## Implementation Steps

### 1. Choose a Source

Start with open data sources (government portals) as they are:
- Legal to use
- Well-documented
- Often have APIs
- Free to access

### 2. Study the Source

- Check if they have an official API
- Review robots.txt
- Check Terms of Service
- Identify data structure

### 3. Create Parser File

Create a new file in `parsers/` directory:

```go
package parsers

import (
    "context"
    "pricemap-go/models"
)

type YourParser struct {
    *BaseParser
}

func NewYourParser() *YourParser {
    return &YourParser{
        BaseParser: NewBaseParser("https://example.com"),
    }
}

func (yp *YourParser) Name() string {
    return "your_parser"
}

func (yp *YourParser) Parse(ctx context.Context) ([]models.Property, error) {
    // Implementation here
    return nil, nil
}
```

### 4. Implement Parse Method

Key steps:
1. Fetch data (use `BaseParser.Fetch()` for HTTP requests)
2. Parse response (HTML, JSON, XML, CSV)
3. Extract property data
4. Create `models.Property` objects
5. Return slice of properties

### 5. Register Parser

Add to `services/scraper.go`:

```go
parsers: []parsers.Parser{
    parsers.NewExampleParser(),
    parsers.NewYourParser(), // Add here
},
```

## Example: Government Open Data Parser

```go
func (gp *GovernmentParser) Parse(ctx context.Context) ([]models.Property, error) {
    var properties []models.Property
    
    // Fetch JSON data from API
    url := "https://data.gov/api/properties"
    body, err := gp.Fetch(ctx, url)
    if err != nil {
        return nil, err
    }
    defer body.Close()
    
    // Parse JSON
    var data struct {
        Properties []struct {
            Address string  `json:"address"`
            Price   float64 `json:"price"`
            Lat     float64 `json:"latitude"`
            Lng     float64 `json:"longitude"`
        } `json:"properties"`
    }
    
    if err := json.NewDecoder(body).Decode(&data); err != nil {
        return nil, err
    }
    
    // Convert to Property models
    for _, p := range data.Properties {
        properties = append(properties, models.Property{
            Source:    gp.Name(),
            Address:   p.Address,
            Price:     p.Price,
            Latitude:  p.Lat,
            Longitude: p.Lng,
            ScrapedAt: time.Now(),
            IsActive:  true,
            Currency:  "USD",
        })
    }
    
    return properties, nil
}
```

## Example: Web Scraping Parser

```go
func (wp *WebParser) Parse(ctx context.Context) ([]models.Property, error) {
    var properties []models.Property
    
    url := fmt.Sprintf("%s/listings", wp.baseURL)
    body, err := wp.Fetch(ctx, url)
    if err != nil {
        return nil, err
    }
    defer body.Close()
    
    doc, err := goquery.NewDocumentFromReader(body)
    if err != nil {
        return nil, err
    }
    
    doc.Find(".property-item").Each(func(i int, s *goquery.Selection) {
        property := wp.parseProperty(s)
        if property != nil {
            properties = append(properties, *property)
        }
    })
    
    return properties, nil
}

func (wp *WebParser) parseProperty(s *goquery.Selection) *models.Property {
    // Extract data from HTML elements
    // ...
    return &models.Property{...}
}
```

## Best Practices

### 1. Error Handling

```go
if err != nil {
    log.Printf("Error in %s: %v", parser.Name(), err)
    // Don't fail entire scraping, continue with other sources
    return nil, err
}
```

### 2. Rate Limiting

```go
// Add delays between requests
time.Sleep(1 * time.Second)
```

### 3. Data Validation

```go
// Validate required fields
if property.Price <= 0 || property.Latitude == 0 {
    return nil // Skip invalid properties
}
```

### 4. Geocoding

If address is available but coordinates are not:

```go
// Use geocoding service (OpenCage, Google Maps, etc.)
lat, lng, err := geocode(property.Address)
if err == nil {
    property.Latitude = lat
    property.Longitude = lng
}
```

### 5. Currency Handling

```go
// Normalize currency
if currency == "€" {
    property.Currency = "EUR"
} else if currency == "$" {
    property.Currency = "USD"
}
```

## Testing Parsers

Create test files:

```go
// parsers/your_parser_test.go
func TestYourParser(t *testing.T) {
    parser := NewYourParser()
    ctx := context.Background()
    
    properties, err := parser.Parse(ctx)
    if err != nil {
        t.Fatalf("Parse failed: %v", err)
    }
    
    if len(properties) == 0 {
        t.Error("No properties parsed")
    }
    
    // Validate property data
    for _, p := range properties {
        if p.Price <= 0 {
            t.Errorf("Invalid price: %f", p.Price)
        }
        if p.Latitude == 0 || p.Longitude == 0 {
            t.Errorf("Missing coordinates")
        }
    }
}
```

## Common Challenges

### 1. Anti-Scraping Measures

- Use proper User-Agent headers
- Implement delays
- Rotate IP addresses (if needed)
- Use headless browsers for JavaScript-heavy sites

### 2. Dynamic Content

- Use Selenium/Playwright for JavaScript-rendered content
- Look for API endpoints that sites use internally

### 3. Data Quality

- Validate all fields
- Handle missing data gracefully
- Normalize formats (dates, prices, etc.)

### 4. Legal Compliance

- Always check Terms of Service
- Respect robots.txt
- Don't overload servers
- Consider using official APIs when available

## Next Steps

1. Start with one open data source (e.g., Data.gov)
2. Implement and test the parser
3. Add to scraper service
4. Monitor and improve
5. Add more sources incrementally

For a complete list of available sources, see [DATA_SOURCES.md](DATA_SOURCES.md).

