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

## Development

To start the development server:

```bash
cd frontend
npm install
npm start
```

The application will be available at `http://localhost:3000` and will proxy API requests to `http://localhost:8080` where the backend should be running.

## Building for Production

To build the application for production:

```bash
npm run build
```

This will create a `build` directory with optimized production files.
