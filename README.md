# YouTube Downloader

A web application for downloading YouTube videos with a web interface and CLI.

## Quick Start

### Using Docker (Recommended)

1. Clone the repository:

   ```bash
   git clone https://github.com/yourusername/youtube-downloader.git
   cd youtube-downloader
   ```

2. Copy and configure environment:

   ```bash
   cp .env.example .env
   # Edit .env if needed
   ```

3. Start the application:
   ```bash
   make build
   docker-compose up
   ```

Access the web interface at http://localhost:8080

### Using CLI

Build and run the CLI:

```bash
go build -o youtube-dl cmd/cli/main.go
./youtube-dl -url "https://www.youtube.com/watch?v=..."
```

## Development Setup

### Backend Development

1. Start the backend with hot-reloading:

   ```bash
   make dev-backend
   ```

2. The API will be available at http://localhost:8080

### Frontend Development

1. Start the frontend with hot-reloading:

   ```bash
   make dev-frontend
   ```

2. The frontend will be available at http://localhost:3000

## Environment Configuration

Key environment variables:

```env
# Server Configuration
PORT=8080                  # Server port
ENV=development            # Environment (development/production)
BASE_URL=http://localhost:8080  # Base URL for download links

# Storage
OUTPUT_DIR=/app/downloads  # Video storage directory
TASK_RETENTION=24h        # How long to keep videos

# Redis
REDIS_ADDR=redis:6379     # Redis server address
```

## API Endpoints

1. Download Video:

   ```bash
   curl -X POST http://localhost:8080/api/download \
     -H "Content-Type: application/json" \
     -d '{"url": "https://www.youtube.com/watch?v=..."}'
   ```

2. Check Status:

   ```bash
   curl http://localhost:8080/api/tasks/{task_id}
   ```

3. Download Video:
   ```bash
   curl http://localhost:8080/videos/{task_id}
   ```

## Architecture

- **Frontend**: React.js
- **Backend**: Go with Chi router
- **Queue**: Redis with Asynq for concurrent task management
- **Video Processing**: yt-dlp

### Concurrent Downloads

The application uses Asynq for background job processing, allowing you to:

- Download multiple videos simultaneously
- Monitor download progress in real-time
- Queue downloads when the system is busy
- Automatically retry failed downloads

Each download runs as a separate task in the queue, with status updates available through the API.

## Monitoring

Access the task queue dashboard at http://localhost:8080/monitoring

## Deployment

1. Build production images:

   ```bash
   make build
   ```

2. Deploy:
   ```bash
   docker-compose up -d
   ```

For production, set appropriate environment variables in `.env`, especially:

- `BASE_URL` for your public domain
- `TASK_RETENTION` for video cleanup
- `ENV=production`
