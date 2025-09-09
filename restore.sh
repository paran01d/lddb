#!/bin/bash

# LDDB SQLite Database Restore Script
# ===================================

set -e

CONTAINER_NAME="lddb-app"
BACKUP_DIR="./backups"

echo "🔄 LDDB Database Restore"
echo "======================="

# Check arguments
if [ $# -eq 0 ]; then
    echo "Usage: $0 <backup_file>"
    echo ""
    echo "Available backups in ${BACKUP_DIR}/:"
    ls -lah "$BACKUP_DIR"/lddb_backup_*.db 2>/dev/null || echo "   No backups found"
    echo ""
    echo "Example: $0 ${BACKUP_DIR}/lddb_backup_20250909_173000.db"
    exit 1
fi

BACKUP_FILE="$1"

# Check if backup file exists
if [ ! -f "$BACKUP_FILE" ]; then
    echo "❌ Backup file not found: $BACKUP_FILE"
    exit 1
fi

# Check if container is running
if ! docker ps | grep -q "$CONTAINER_NAME"; then
    echo "❌ Container '$CONTAINER_NAME' is not running"
    echo "   Start the application first: docker compose up -d"
    exit 1
fi

echo "📋 Backup file: $BACKUP_FILE"
echo "   Size: $(du -h "$BACKUP_FILE" | cut -f1)"

# Show backup contents summary
echo ""
echo "📊 Backup Contents:"
TOTAL_COUNT=$(sqlite3 "$BACKUP_FILE" "SELECT COUNT(*) FROM laserdiscs;")
WATCHED_COUNT=$(sqlite3 "$BACKUP_FILE" "SELECT COUNT(*) FROM laserdiscs WHERE watched = 1;")
UNWATCHED_COUNT=$((TOTAL_COUNT - WATCHED_COUNT))

echo "   Total LaserDiscs: $TOTAL_COUNT"
echo "   Watched: $WATCHED_COUNT"  
echo "   Unwatched: $UNWATCHED_COUNT"

# Confirmation
echo ""
read -p "⚠️  This will replace your current collection. Continue? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ Restore cancelled"
    exit 0
fi

echo ""
echo "🔄 Stopping container for safe restore..."
docker compose stop lddb

echo "📥 Restoring database..."
# Copy backup file to container volume
docker cp "$BACKUP_FILE" "${CONTAINER_NAME}:/app/data/collection.db"

echo "🚀 Starting container..."
docker compose start lddb

# Wait for container to be ready
echo "⏳ Waiting for application to start..."
sleep 5

# Verify restore
if curl -s http://localhost:8090/api/collection >/dev/null 2>&1; then
    echo "✅ Restore completed successfully!"
    echo "   Application is running at: http://localhost:8090"
    echo "   HTTPS: https://cooee.mankies.com"
else
    echo "⚠️  Restore completed but application may need a moment to start"
    echo "   Check status: docker compose ps"
fi