# Video Streaming Backend

A Go Fiber-based backend for video streaming application.

## Features

- Video upload and storage
- Video streaming
- Video listing
- Video deletion

## Setup

1. Install Go (1.16 or later)
2. Install dependencies: `go get .`
3. Run the server: `go run main.go`

The server will start on port 8080.

## API Endpoints

- `GET /api/videos` - List all videos
- `GET /api/videos/:id` - Get video details
- `POST /api/videos` - Upload a video (multipart/form-data with 'video' field)
- `DELETE /api/videos/:id` - Delete a video

Videos are served from `/videos/:filename` endpoint. 