# Video Streaming Platform

A full-stack video streaming platform with Next.js frontend and Go Fiber backend.

## Project Structure

- `frontend/` - Next.js frontend
- `backend/` - Go Fiber backend

## Features

- Video uploading and streaming
- Progress bar for uploads
- Video listing
- Video deletion
- REST API
- **Video Transcoding** - Convert videos to streaming-friendly formats (MP4, HLS, DASH)
- **Adaptive Streaming** - HLS and DASH for adaptive bitrate streaming
- **Auto Replay** - Automatically restart videos when they finish playing
- **Transcoding Progress** - Real-time progress indicator for video transcoding operations

## Getting Started

### Prerequisites

- [FFmpeg](https://ffmpeg.org/download.html) - Required for video transcoding
  - See [FFmpeg Setup Guide](backend/ffmpeg-setup.md) for installation instructions

### Backend Setup

1. Make sure Go is installed (1.16 or later)
2. Navigate to the backend directory: `cd backend`
3. Install dependencies: `go mod tidy`
4. Run the server: `go run main.go`

The backend will start on port 8080.

### Frontend Setup

1. Make sure Node.js is installed (v18 or later recommended)
2. Navigate to the frontend directory: `cd frontend`
3. Install dependencies: `npm install`
4. Start the development server: `npm run dev`

The frontend will start on port 3000.

## Usage

1. Open `http://localhost:3000` in your browser
2. Upload videos using the upload button
3. Videos will be listed in the sidebar
4. Click on a video to play it
5. Use the delete button to remove videos
6. Click "Transcode Video" to convert the video to different formats:
   - **MP4**: Standard video format with different resolutions
   - **HLS**: HTTP Live Streaming for Apple devices and adaptive streaming
   - **DASH**: Dynamic Adaptive Streaming over HTTP for cross-platform adaptive streaming

## Video Transcoding

The platform supports transcoding videos to different formats and qualities:

### Supported Formats

- **MP4** - Standard video format compatible with all browsers
- **HLS** - HTTP Live Streaming (creates .m3u8 playlist and .ts segments)
- **DASH** - Dynamic Adaptive Streaming over HTTP (creates .mpd manifest and mp4 segments)

### Transcoding Options

- **Resolution**: 240p, 360p, 480p, 720p (HD), 1080p (Full HD), 1440p (2K), 2160p (4K)
- **Bitrate**: Very Low (500k), Low (1000k), Medium (2000k), High (4000k), Very High (8000k), Ultra (16000k)

### Progress Monitoring

The platform provides real-time progress updates during video transcoding operations, allowing users to monitor the conversion process. The progress bar shows the percentage completion of the current transcoding task.

## Technologies Used

### Frontend
- Next.js
- TypeScript
- Tailwind CSS
- HLS.js for HLS streaming

### Backend
- Go
- Fiber
- Go modules
- FFmpeg for video processing 