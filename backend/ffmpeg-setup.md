# FFmpeg Setup Guide

This guide helps you set up FFmpeg, which is required for the video transcoding functionality in our application.

## What is FFmpeg?

FFmpeg is a free and open-source software project consisting of a suite of libraries and programs for handling video, audio, and other multimedia files and streams.

## Installation

### Windows

1. Download FFmpeg from the official website: https://ffmpeg.org/download.html 
   - Or direct link to Windows builds: https://www.gyan.dev/ffmpeg/builds/

2. Download the "ffmpeg-release-essentials.zip" file

3. Extract the ZIP file to a location on your computer (e.g., `C:\ffmpeg`)

4. Add FFmpeg to your system PATH:
   - Right-click on "This PC" or "My Computer" and select "Properties"
   - Click on "Advanced system settings"
   - Click on "Environment Variables"
   - Under "System variables", find the "Path" variable, select it and click "Edit"
   - Click "New" and add the path to the `bin` folder (e.g., `C:\ffmpeg\bin`)
   - Click "OK" to close all dialogs

5. Verify the installation by opening a new Command Prompt and typing:
   ```
   ffmpeg -version
   ```

### macOS

Using Homebrew:

```
brew install ffmpeg
```

Verify the installation:

```
ffmpeg -version
```

### Linux (Ubuntu/Debian)

```
sudo apt update
sudo apt install ffmpeg
```

Verify the installation:

```
ffmpeg -version
```

## Supported Transcoding Formats

Our application supports transcoding videos to the following formats:

1. **MP4** - Standard video format compatible with almost all devices and browsers
2. **HLS (HTTP Live Streaming)** - Adaptive streaming protocol developed by Apple
3. **DASH (Dynamic Adaptive Streaming over HTTP)** - Adaptive streaming protocol by MPEG

## Transcoding Options

You can customize the following options when transcoding a video:

- **Format**: Choose between MP4, HLS, or DASH
- **Resolution**: 240p, 360p, 480p, 720p (HD), 1080p (Full HD), 1440p (2K), or 2160p (4K)
- **Bitrate**: Very Low (500k), Low (1000k), Medium (2000k), High (4000k), Very High (8000k), or Ultra (16000k)

## Progress Monitoring

The application provides real-time progress updates during the transcoding process. This allows you to:

- Track the percentage completion of your transcoding jobs
- Estimate remaining time for lengthy transcoding operations
- Get immediate feedback if the transcoding process fails

The progress monitoring works by:
1. Analyzing the FFmpeg output in real-time
2. Calculating the percentage based on the video duration and current processing time
3. Displaying the progress through the user interface

This feature is particularly useful when transcoding large files or when selecting high-resolution options that require more processing time.

## Troubleshooting

If you encounter issues with transcoding, check the following:

1. Verify FFmpeg is installed correctly by running `ffmpeg -version` in your terminal
2. Ensure the upload directory has proper write permissions
3. Check the server logs for detailed error messages
4. Make sure the video file you're trying to transcode is not corrupted

For more information about FFmpeg, visit the official documentation: https://ffmpeg.org/documentation.html 