# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- **Global Coverage**: Parsers now iterate over all cities in each country
- **Sale & Rent Support**: All parsers support both sale and rental properties
- **New Parsers**:
  - ZillowParser (USA) - 30+ cities
  - IdealistaParser (Spain) - 20+ cities
- **Crime Data Integration**: Real crime data from UK Police API and other sources
- **Education Data Integration**: School ratings and education statistics
- **Transportation Service**: GTFS data parsing and transport score calculation
- **Caching Layer**: In-memory cache with TTL for performance
- **Metrics Service**: Track parsing performance, errors, and statistics
- **API Metrics Endpoint**: `/api/v1/metrics` for system monitoring
- **Currency Converter**: Convert prices to USD for comparison
- **Property Validation**: Validate properties before saving
- **Docker Support**: Dockerfile and docker-compose.yml for easy deployment
- **CI/CD**: GitHub Actions workflow for testing and linting
- **Makefile**: Convenient commands for common tasks
- **Middleware**: Rate limiting, CORS, logging
- **City Service**: Get cities from OpenStreetMap or curated lists
- **Geocoding Service**: OpenCage and Nominatim support

### Changed
- All parsers now support multiple cities
- Factors calculation uses real APIs where available
- Improved error handling and logging
- Better rate limiting between requests

### Fixed
- Import cycle issues
- Missing dependencies
- Compilation errors

## [0.1.0] - Initial Release

### Added
- Basic project structure
- CianParser for Russia
- RightmoveParser for UK
- OpenDataParser for government data
- API server with heatmap endpoints
- Frontend with interactive map
- Database models and migrations
- Factor calculation (crime, transport, education, infrastructure)
- Documentation

