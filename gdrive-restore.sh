#!/bin/bash

# LDDB Google Drive Restore Script
# ================================

set -e

BACKUP_DIR="./backups"
GDRIVE_FOLDER="LDDB_Backups"

echo "☁️  LDDB Google Drive Restore"
echo "============================"

# Check if gdrive CLI is installed
if ! command -v gdrive &> /dev/null; then
    echo "❌ Google Drive CLI tool 'gdrive' is not installed"
    echo "   Install first: see gdrive-backup.sh for instructions"
    exit 1
fi

# Check if LDDB_Backups folder exists
FOLDER_ID=$(gdrive files list --query "name='$GDRIVE_FOLDER' and mimeType='application/vnd.google-apps.folder'" --skip-header | awk '{print $1}' | head -1)

if [ -z "$FOLDER_ID" ]; then
    echo "❌ LDDB_Backups folder not found in Google Drive"
    echo "   Create a backup first: ./gdrive-backup.sh"
    exit 1
fi

echo "☁️  Available backups in Google Drive:"
echo ""

# List available backups
gdrive files list --parent "$FOLDER_ID"

echo ""

if [ $# -eq 0 ]; then
    echo "Usage: $0 <google_drive_file_id>"
    echo ""
    echo "Example:"
    echo "   1. Copy the File ID from the list above"
    echo "   2. Run: $0 1abc...xyz"
    exit 1
fi

FILE_ID="$1"

# Create backup directory
mkdir -p "$BACKUP_DIR"

echo "📥 Downloading backup from Google Drive..."
echo "   File ID: $FILE_ID"

# Download the backup file (gdrive downloads with original filename)
gdrive files download "$FILE_ID" --destination "$BACKUP_DIR" --overwrite

# Find the downloaded file (gdrive uses original filename)
DOWNLOAD_FILE=$(find "$BACKUP_DIR" -name "*.db" -newer "$BACKUP_DIR" 2>/dev/null | head -1)
if [ -z "$DOWNLOAD_FILE" ]; then
    # Fallback: look for most recent .db file
    DOWNLOAD_FILE=$(ls -t "$BACKUP_DIR"/*.db 2>/dev/null | head -1)
fi

if [ $? -eq 0 ] && [ -f "$DOWNLOAD_FILE" ]; then
    echo "✅ Downloaded successfully!"
    echo ""
    
    # Show backup contents
    echo "📊 Backup Contents:"
    TOTAL_COUNT=$(sqlite3 "$DOWNLOAD_FILE" "SELECT COUNT(*) FROM laserdiscs;" 2>/dev/null || echo "0")
    WATCHED_COUNT=$(sqlite3 "$DOWNLOAD_FILE" "SELECT COUNT(*) FROM laserdiscs WHERE watched = 1;" 2>/dev/null || echo "0")
    UNWATCHED_COUNT=$((TOTAL_COUNT - WATCHED_COUNT))
    
    echo "   Total LaserDiscs: $TOTAL_COUNT"
    echo "   Watched: $WATCHED_COUNT"
    echo "   Unwatched: $UNWATCHED_COUNT"
    echo ""
    
    # Use existing restore script
    echo "🔄 Proceeding with local restore..."
    ./restore.sh "$DOWNLOAD_FILE"
    
else
    echo "❌ Failed to download backup from Google Drive"
    exit 1
fi