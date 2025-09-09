#!/bin/bash

# LDDB SQLite Database Backup Script
# ==================================

set -e

# Configuration
BACKUP_DIR="./backups"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="lddb_backup_${TIMESTAMP}.db"
CONTAINER_NAME="lddb-app"

echo "🗃️  LDDB Database Backup"
echo "======================="

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

# Check if container is running
if ! docker ps | grep -q "$CONTAINER_NAME"; then
    echo "❌ Container '$CONTAINER_NAME' is not running"
    echo "   Start the application first: docker compose up -d"
    exit 1
fi

echo "📋 Backing up database from container..."

# Copy database from container to host
docker cp "${CONTAINER_NAME}:/app/data/collection.db" "${BACKUP_DIR}/${BACKUP_FILE}"

if [ $? -eq 0 ]; then
    echo "✅ Backup created successfully!"
    echo "   File: ${BACKUP_DIR}/${BACKUP_FILE}"
    echo "   Size: $(du -h "${BACKUP_DIR}/${BACKUP_FILE}" | cut -f1)"
    
    # Show backup contents summary
    echo ""
    echo "📊 Backup Summary:"
    TOTAL_COUNT=$(sqlite3 "${BACKUP_DIR}/${BACKUP_FILE}" "SELECT COUNT(*) FROM laserdiscs;")
    WATCHED_COUNT=$(sqlite3 "${BACKUP_DIR}/${BACKUP_FILE}" "SELECT COUNT(*) FROM laserdiscs WHERE watched = 1;")
    UNWATCHED_COUNT=$((TOTAL_COUNT - WATCHED_COUNT))
    
    echo "   Total LaserDiscs: $TOTAL_COUNT"
    echo "   Watched: $WATCHED_COUNT"
    echo "   Unwatched: $UNWATCHED_COUNT"
    
    echo ""
    echo "📁 All backups in ${BACKUP_DIR}/:"
    ls -lah "$BACKUP_DIR"/lddb_backup_*.db 2>/dev/null | tail -5
    
else
    echo "❌ Backup failed!"
    exit 1
fi