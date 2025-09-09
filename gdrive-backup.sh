#!/bin/bash

# LDDB Google Drive Backup Script
# ===============================

set -e

BACKUP_DIR="./backups"
GDRIVE_FOLDER="LDDB_Backups"
GOOGLE_ACCOUNT="paran01d@gmail.com"

echo "☁️  LDDB Google Drive Backup"
echo "============================"

# Check if gdrive CLI is installed
if ! command -v gdrive &> /dev/null; then
    echo "❌ Google Drive CLI tool 'gdrive' is not installed"
    echo ""
    echo "📥 Install gdrive:"
    echo "   wget -O gdrive 'https://github.com/glotlabs/gdrive/releases/latest/download/gdrive_linux-x64.tar.gz'"
    echo "   tar -xzf gdrive_linux-x64.tar.gz"
    echo "   sudo mv gdrive /usr/local/bin/"
    echo "   chmod +x /usr/local/bin/gdrive"
    echo ""
    echo "🔐 Then authenticate:"
    echo "   gdrive account add --service-account"
    echo "   # Or use: gdrive account add for OAuth"
    exit 1
fi

# Create local backup first
echo "📋 Creating local backup..."
./backup.sh

# Get the latest backup file
LATEST_BACKUP=$(ls -t "$BACKUP_DIR"/lddb_backup_*.db | head -1)

if [ ! -f "$LATEST_BACKUP" ]; then
    echo "❌ No backup file found"
    exit 1
fi

echo "☁️  Uploading to Google Drive..."
echo "   Account: $GOOGLE_ACCOUNT"
echo "   File: $(basename "$LATEST_BACKUP")"
echo "   Size: $(du -h "$LATEST_BACKUP" | cut -f1)"

# Check if LDDB_Backups folder exists, create if not
FOLDER_ID=$(gdrive files list --query "name='$GDRIVE_FOLDER' and mimeType='application/vnd.google-apps.folder'" --skip-header | awk '{print $1}' | head -1)

if [ -z "$FOLDER_ID" ]; then
    echo "📁 Creating LDDB_Backups folder..."
    FOLDER_ID=$(gdrive files mkdir "$GDRIVE_FOLDER" --print-only-id)
fi

# Upload backup to Google Drive
UPLOAD_RESULT=$(gdrive files upload "$LATEST_BACKUP" --parent "$FOLDER_ID" --print-only-id)

if [ $? -eq 0 ]; then
    echo "✅ Backup uploaded successfully to Google Drive!"
    echo "   Google Drive File ID: $UPLOAD_RESULT"
    echo "   Folder: $GDRIVE_FOLDER"
    
    # List recent backups in Google Drive
    echo ""
    echo "☁️  Recent backups in Google Drive:"
    gdrive files list --query "parents in '$FOLDER_ID'" --max 5
    
else
    echo "❌ Failed to upload backup to Google Drive"
    exit 1
fi

echo ""
echo "💡 To restore from Google Drive:"
echo "   1. List backups: gdrive files list --query \"parents in '$FOLDER_ID'\""
echo "   2. Download: gdrive files download <FILE_ID> --path ./backups/"
echo "   3. Restore: ./restore.sh ./backups/downloaded_backup.db"