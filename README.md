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

### Using Docker (Recommended)

1. Clone the repository
2. Start with Docker Compose: `docker-compose up -d`
3. Access securely at **https://cooee.mankies.com** (or http://localhost:8090 for local testing)
4. Start scanning LaserDisc barcodes with secure camera access!

### Manual Installation

1. Clone the repository
2. Install Go dependencies: `go mod download`  
3. Run the application: `go run cmd/server/main.go`
4. Open http://localhost:8082 (or whatever port it finds) in your browser

### Docker Commands

```bash
# Start the application
docker compose up -d

# Stop the application  
docker compose down

# View logs
docker compose logs -f

# Rebuild after changes
docker compose up -d --build
```

### Backup and Restore

**Create Backup:**
```bash
./backup.sh
```
- Creates timestamped backup in `./backups/` directory
- Shows collection summary (total, watched, unwatched counts)
- SQLite database file with all metadata and cover images

**Restore from Backup:**
```bash
./restore.sh ./backups/lddb_backup_20250909_173000.db
```
- Lists available backup files if no argument provided
- Shows backup contents before restoring
- Safely stops/starts application during restore
- Confirms restoration success

### Google Drive Cloud Backup

**Setup (One-time):**
```bash
# Install gdrive CLI
wget -O gdrive.tar.gz 'https://github.com/glotlabs/gdrive/releases/latest/download/gdrive_linux-x64.tar.gz'
tar -xzf gdrive.tar.gz
sudo mv gdrive /usr/local/bin/
chmod +x /usr/local/bin/gdrive

# Authenticate with your Google account
gdrive account add
```

**Cloud Backup:**
```bash
./gdrive-backup.sh
```
- Creates local backup then uploads to Google Drive
- Stores in "LDDB_Backups" folder
- Shows upload confirmation and file ID

**Cloud Restore:**
```bash
./gdrive-restore.sh <google_drive_file_id>
```
- Lists available cloud backups
- Downloads and restores specified backup
- Complete cloud-to-local restoration workflow

### SSL Configuration

**For Local Network Access:**
1. Add to your `/etc/hosts` file: `127.0.0.1 grave.local`  
2. Access at **https://grave.local** with automatic Let's Encrypt SSL
3. Camera access will work securely on mobile devices

**For Production Deployment:**
1. Point your domain DNS A record to your server's public IP
2. Update `Caddyfile` with your domain name
3. Caddy automatically handles Let's Encrypt certificates
4. HTTPS enforced for secure camera API access

**Why HTTPS is Important:**
- Modern browsers require HTTPS for camera/microphone access
- Mobile devices need secure context for barcode scanning
- Let's Encrypt provides free, automatic SSL certificates
- Enhanced security for your LaserDisc collection data

## Development

This project follows conventional commits and is organized with a clean architecture separating concerns into distinct packages.

## License

MIT License - Personal use project