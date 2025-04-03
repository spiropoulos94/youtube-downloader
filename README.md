# YouTube Downloader

A web application for downloading YouTube videos.

## Features

- Download YouTube videos
- View download status
- Monitor task queue
- Responsive UI

## Running the Application

There are two ways to run the application:

### 1. Production Mode

To run the complete application in production mode:

1. Copy the environment file:

   ```
   cp .env.example .env
   ```

2. Build and run with Docker Compose:

   ```
   make build
   docker-compose up
   ```

3. Access the application at http://localhost:8080

### 2. Development Mode

For active development with hot-reloading:

#### Backend (with hot reloading)

```bash
cp .env.example .env
make dev-backend
```

The backend API will be available at http://localhost:8080

#### Frontend (with hot reloading)

```bash
make dev-frontend
```

The frontend will be available at http://localhost:3000

## Architecture

- **Frontend**: React.js application
- **Backend**: Go server with Chi router
- **Database**: Redis for task queue and status
- **Video Processing**: yt-dlp for downloading videos

## Monitoring

A monitoring dashboard for the task queue is available at http://localhost:8080/monitoring

## API Usage

- YouTube API for video information
- Redis for task queue management
- File system for video storage

## Tech Stack

- **Frontend**: React.js
- **Backend**: Go (Golang)
- **Queue System**: Redis
- **Docker**: Used for deployment and development

## Environment Variables

You can configure the application using environment variables:

```env
# Server Configuration
PORT=8080                  # The port the server will listen on
ENV=development            # Environment (development/production)
BASE_URL=http://localhost:8080  # Base URL for generating absolute URLs (e.g., download links)

# File Storage
OUTPUT_DIR=/app/downloads  # Directory where videos will be saved

# Redis Configuration
REDIS_ADDR=redis:6379      # Redis server address

# Data Retention
TASK_RETENTION=24h         # How long to keep videos before cleanup (e.g., 24h, 7d)

# Optional: Logging
LOG_LEVEL=info             # Logging level (info, debug, error)
```

## Prerequisites

- Go 1.24 or higher
- yt-dlp installed on your system:

  ```bash
  # macOS
  brew install yt-dlp

  # Linux
  sudo apt install yt-dlp
  # or
  sudo pip install yt-dlp
  ```

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/yourusername/youtube-downloader.git
   cd youtube-downloader
   ```

2. Copy the example environment file:

   ```bash
   cp .env.example .env
   ```

3. (Optional) Modify the `.env` file with your preferred settings.

## Development with Hot Reloading

### Frontend Hot Reloading

The frontend automatically includes hot reloading through React's development server when running with `make dev-frontend`.

### Backend Hot Reloading with Air

For backend development, there are two options for hot reloading with Air:

#### Option 1: Using Docker (Recommended)

When using `make dev-backend`, Air is already configured in the Docker container. The system uses:

- Dockerfile.dev which installs Air and necessary dependencies
- dev-start.sh script that sets up the Python virtual environment and runs Air
- .air.toml configuration that watches your Go files and rebuilds on changes

No additional setup is needed as everything is handled automatically.

#### Option 2: Direct Local Development

If you prefer to run the backend directly on your machine:

1. Install Air:

   ```bash
   # Using Go
   go install github.com/cosmtrek/air@latest

   # Or using Homebrew on macOS
   brew install air
   ```

2. The project includes an `.air.toml` configuration file

3. Run Air:
   ```bash
   air
   ```

Both options will automatically rebuild and restart your Go server whenever you make changes to your code.

## Usage

### Command Line Interface

Download videos directly from the terminal:

```bash
# Build the CLI
go build -o youtube-dl cmd/cli/main.go

# Run the CLI
./youtube-dl -url "https://www.youtube.com/watch?v=..." -output downloads
```

Options:

- `-url`: YouTube video URL (required)
- `-output`: Output directory (default: "downloads")

### Web Server

Run the HTTP server to download videos via API:

```bash
# Build the server
go build -o server cmd/server/main.go

# Run the server
./server -port 8080 -output downloads
```

Options:

- `-port`: Server port (default: "8080")
- `-output`: Output directory (default: "downloads")
- `-env`: Environment (development/production, default: "development")

#### API Endpoints

1. Download Video

   ```bash
   curl -X POST http://localhost:8080/api/download \
     -H "Content-Type: application/json" \
     -d '{"url": "https://www.youtube.com/watch?v=..."}'
   ```

   Response:

   ```json
   {
     "success": true,
     "data": {
       "task_id": "550e8400-e29b-41d4-a716-446655440000"
     }
   }
   ```

2. Check Download Status

   ```bash
   curl http://localhost:8080/api/tasks/550e8400-e29b-41d4-a716-446655440000
   ```

   Response:

   ```json
   {
     "success": true,
     "data": {
       "status": "completed",
       "file_path": "/path/to/downloaded/video.mp4",
       "download_url": "http://localhost:8080/videos/550e8400-e29b-41d4-a716-446655440000",
       "title": "Video Title",
       "thumbnail_url": "https://i.ytimg.com/vi/video_id/maxresdefault.jpg",
       "duration": "5:32"
     }
   }
   ```

3. Health Check
   ```bash
   curl http://localhost:8080/api/health
   ```

## Cleaning Up

To clean up all containers and volumes:

```bash
make clean
```

## Deployment

For deployment to production:

1. Build the production Docker images:

   ```bash
   make build
   ```

2. You can deploy the built images to any container hosting service (Docker Swarm, Kubernetes, etc.)

3. For simple deployments:
   ```bash
   docker-compose up -d
   ```

## URL Configuration

The application uses the `BASE_URL` environment variable to generate absolute URLs for video downloads. This ensures correct URL generation in various deployment scenarios:

- **Development**: `BASE_URL=http://localhost:8080`
- **Production**: `BASE_URL=https://yourdomain.com` or `BASE_URL=https://api.yourdomain.com`

If your application is deployed at a subpath, include it in the BASE_URL:

```
BASE_URL=https://yourdomain.com/youtube-downloader
```

This configuration is especially important when:

- The application is behind a reverse proxy or load balancer
- You're using HTTPS terminated at the proxy level
- The internal server URL differs from the public-facing URL
