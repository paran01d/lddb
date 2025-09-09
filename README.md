# LDDB - LaserDisc Collection Manager

A web-based LaserDisc collection management application with barcode scanning, automatic metadata lookup, and collection management features.

## Features

- 📱 **Mobile-friendly barcode scanner** using device camera
- 🔍 **Automatic metadata lookup** by scraping lddb.com
- 📚 **Complete collection management** (CRUD operations)
- 🎲 **Random unwatched title selector** for movie night decisions
- 📱 **Responsive design** works on desktop and mobile
- 🔄 **Progressive Web App** with offline capability
- 🔎 **Search and filtering** for large collections

## Technology Stack

- **Backend**: Go with Gin framework
- **Database**: SQLite with GORM ORM
- **Frontend**: Vanilla JavaScript + CSS (no frameworks)
- **Barcode Scanning**: QuaggaJS camera API
- **Web Scraping**: Colly for lddb.com data extraction

## Quick Start

1. Clone the repository
2. Install Go dependencies: `go mod download`
3. Run the application: `go run cmd/server/main.go`
4. Open http://localhost:8080 in your browser

## Development

This project follows conventional commits and is organized with a clean architecture separating concerns into distinct packages.

## License

MIT License - Personal use project