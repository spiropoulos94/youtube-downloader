# YouTube Video Downloader

A Go application that allows you to download YouTube videos either through a command-line interface or a web server. The application always downloads videos in the best available quality.

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
       "file_path": "/path/to/downloaded/video.mp4"
     }
   }
   ```

3. Health Check
   ```bash
   curl http://localhost:8080/api/health
   ```

## Environment Variables

You can configure the application using environment variables or command-line flags:

```env
PORT=8080
OUTPUT_DIR=downloads
ENV=development
```

Command-line flags take precedence over environment variables.

## Development

Run the server in development mode:

```bash
go run cmd/server/main.go
```

Run the CLI:

```bash
go run cmd/cli/main.go -url "https://www.youtube.com/watch?v=..."
```

## License

MIT License
