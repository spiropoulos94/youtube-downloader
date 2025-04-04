#!/bin/sh
set -e

echo "Starting development environment..."

# Create and activate virtual environment
python3 -m venv /app/venv
. /app/venv/bin/activate
pip install --no-cache-dir yt-dlp

# Start air for hot reloading
cd /app
exec air -c .air.toml
