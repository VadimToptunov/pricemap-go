# Quick Start Guide

## 🚀 Server is Running!

Your PriceMap server is now running and ready to use.

### Access Points

- **API Server**: http://localhost:8080
- **API Documentation**: See README.md for all endpoints
- **Frontend**: Open `web/index.html` in your browser (update API key first)

### Available Endpoints

1. **GET /api/v1/stats** - Get statistics
2. **GET /api/v1/properties** - List properties
3. **GET /api/v1/properties/:id** - Get property details
4. **GET /api/v1/heatmap** - Get heatmap data
5. **GET /api/v1/metrics** - Get system metrics

### Next Steps

#### 1. Start Parsing Data

Run the scraper to collect real estate data:

```bash
# One-time scraping
go run cmd/scraper/main.go

# Or use Docker
docker-compose up scraper
```

#### 2. Start Scheduler (Automatic Updates)

Run the scheduler for automatic periodic updates:

```bash
# Manual
go run cmd/scheduler/main.go

# Or use Docker
docker-compose up scheduler
```

#### 3. View Data

Once data is collected, you can:

- View heatmap: http://localhost:8080/api/v1/heatmap?lat_min=55&lat_max=56&lng_min=37&lng_max=38
- View properties: http://localhost:8080/api/v1/properties
- View stats: http://localhost:8080/api/v1/stats
- View metrics: http://localhost:8080/api/v1/metrics

#### 4. Open Frontend

1. Edit `web/index.html` and add your Google Maps API key
2. Open `web/index.html` in your browser
3. The map will load data from the API server

### Docker Commands

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f server
docker-compose logs -f scraper
docker-compose logs -f scheduler

# Stop services
docker-compose down

# Restart services
docker-compose restart
```

### Troubleshooting

#### Server not starting?
- Check if port 8080 is available: `lsof -ti:8080`
- Check database connection in logs
- Verify .env file exists with correct DB settings

#### No data?
- Run the scraper: `go run cmd/scraper/main.go`
- Check parser logs for errors
- Verify parsers are registered in `services/scraper.go`

#### Database issues?
- Check Docker container: `docker ps | grep pricemap-db`
- View logs: `docker logs pricemap-db`
- Restart: `docker-compose restart postgres`

### Useful Commands

```bash
# Build all
make build

# Run tests
make test

# Format code
make fmt

# Clean
make clean
```

### Current Status

✅ **Server**: Running on port 8080
✅ **Database**: PostgreSQL in Docker
✅ **API**: All endpoints available
⏳ **Data**: Run scraper to collect data
⏳ **Frontend**: Update Google Maps API key

Happy parsing! 🎉

