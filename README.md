# YouTube Video Downloader

A Go application that allows you to download YouTube videos either through a command-line interface or a web server.

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
./youtube-dl -url "https://www.youtube.com/watch?v=..." -quality 1080p -output downloads
```

Options:

- `-url`: YouTube video URL (required)
- `-quality`: Video quality (default: "best")
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
     -d '{"url": "https://www.youtube.com/watch?v=...", "quality": "1080p"}'
   ```

2. Health Check
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
