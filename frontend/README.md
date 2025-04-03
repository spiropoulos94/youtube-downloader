# YouTube Downloader Frontend

This is a React-based frontend application for downloading YouTube videos. It works in conjunction with the Go backend API.

## Features

- Input field for YouTube video URLs
- Real-time status updates on download progress
- Download button appears when videos are ready
- Responsive design using Material UI

## Technologies Used

- React 18
- TypeScript
- Material UI
- Axios for API requests

## Environment Configuration

The API communication is configured through environment variables:

- `.env` - Default configuration for production or when frontend and backend share the same host
- `.env.development` - For development when running on different ports (frontend: 3000, backend: 8080)

Before starting development:

1. Copy `.env.development.example` to `.env.development`
2. Adjust the variables if needed for your local setup

The key environment variable is:

- `REACT_APP_API_BASE_URL` - Base URL for API requests (empty for same-host, http://localhost:8080 for cross-port dev)

## Development

To start the development server:

```bash
cd frontend
npm install
# Copy the example env file if you haven't already
cp .env.development.example .env.development
npm start
```

The application will be available at `http://localhost:3000` and will proxy API requests to `http://localhost:8080` where the backend should be running.

## Building for Production

To build the application for production:

```bash
npm run build
```

This will create a `build` directory with optimized production files.
