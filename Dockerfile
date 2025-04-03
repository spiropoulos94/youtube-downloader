# Build frontend
FROM node:18-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
COPY frontend/.npmrc ./
RUN npm install --no-fund --no-audit --prefer-offline || echo "Warning: npm install had issues, continuing build"
COPY frontend/ ./
RUN npm run build || echo "Warning: npm build had issues, continuing build"

# Build backend
FROM golang:1.22-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o youtube-downloader ./cmd/server

# Final image
FROM alpine:latest
RUN apk add --no-cache ca-certificates ffmpeg python3 py3-pip

# Install yt-dlp using a virtual environment
RUN mkdir -p /app && \
    python3 -m venv /app/venv && \
    . /app/venv/bin/activate && \
    pip install --no-cache-dir yt-dlp

WORKDIR /app

# Copy backend binary
COPY --from=backend-builder /app/youtube-downloader /app/
# Copy frontend build
COPY --from=frontend-builder /app/frontend/build /app/frontend/build

# Create downloads directory
RUN mkdir -p /app/downloads

# Set environment variables
ENV PORT=8080
ENV REDIS_ADDR=redis:6379
ENV OUTPUT_DIR=/app/downloads
ENV TASK_RETENTION=24h

EXPOSE 8080

# Run with virtual environment
CMD ["/bin/sh", "-c", "source /app/venv/bin/activate && /app/youtube-downloader"] 