#!/bin/sh
set -e
echo "Starting development environment..."
python3 -m venv /app/venv || true
. /app/venv/bin/activate
pip install --no-cache-dir yt-dlp
cd /app
exec air -c .air.toml
