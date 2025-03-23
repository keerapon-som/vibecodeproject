package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
)

// Define global directory for videos
var uploadsDir = "./uploads/videos"
var transcodedDir = "./uploads/transcoded"

// Map to store transcoding progress
var transcodingProgress = make(map[string]int)

func main() {
	app := fiber.New(fiber.Config{
		BodyLimit: 2000 * 1024 * 1024, // 2GB for video uploads
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000", // Next.js frontend
		AllowHeaders:     "Origin, Content-Type, Accept, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
		ExposeHeaders:    "Content-Length, Content-Type",
		MaxAge:           86400, // 24 hours
	}))

	// Log uploads directory
	log.Printf("Using uploads directory: %s", uploadsDir)
	absPath, err := filepath.Abs(uploadsDir)
	if err == nil {
		log.Printf("Absolute path: %s", absPath)
	}

	// Ensure videos directory exists
	if err := os.MkdirAll(uploadsDir, os.ModePerm); err != nil {
		log.Fatal("Failed to create videos directory:", err)
	}

	// Ensure transcoded directory exists
	if err := os.MkdirAll(transcodedDir, os.ModePerm); err != nil {
		log.Fatal("Failed to create transcoded directory:", err)
	}

	// Static files serving
	app.Static("/videos", uploadsDir)
	app.Static("/transcoded", transcodedDir)

	// Websocket route for transcoding progress
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/transcode/:id", websocket.New(func(c *websocket.Conn) {
		// Get video ID from URL
		videoId := c.Params("id")

		// Send progress updates to the client
		for {
			progress, exists := transcodingProgress[videoId]
			if !exists {
				progress = 0
			}

			if err := c.WriteJSON(fiber.Map{
				"videoId":  videoId,
				"progress": progress,
			}); err != nil {
				log.Println("Error writing to websocket:", err)
				break
			}

			// If processing is complete, break the loop
			if progress >= 100 {
				// Keep connection for a moment so final message is received
				time.Sleep(1 * time.Second)
				break
			}

			time.Sleep(500 * time.Millisecond)
		}
	}))

	// API Routes
	api := app.Group("/api")

	// Video routes
	videos := api.Group("/videos")
	videos.Get("/", getVideos)
	videos.Get("/:id", getVideo)
	videos.Post("/", uploadVideo)
	videos.Delete("/:id", deleteVideo)
	videos.Post("/transcode/:id", transcodeVideo)

	// Progress endpoint for polling
	api.Get("/transcode/progress/:id", func(c *fiber.Ctx) error {
		videoId := c.Params("id")
		progress, exists := transcodingProgress[videoId]
		if !exists {
			progress = 0
		}

		return c.JSON(fiber.Map{
			"videoId":  videoId,
			"progress": progress,
		})
	})

	// Start server
	log.Printf("Starting server on port 8080")
	log.Fatal(app.Listen(":8080"))
}

// checkFFmpeg checks if FFmpeg is installed
func checkFFmpeg() error {
	cmd := exec.Command("ffmpeg", "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("FFmpeg is not installed or not in PATH: %v", err)
	}
	return nil
}

// parseProgress parses the FFmpeg output to extract progress information
func parseProgress(output string, duration float64) int {
	// Regular expression to match the time
	re := regexp.MustCompile(`time=(\d+):(\d+):(\d+\.\d+)`)
	matches := re.FindStringSubmatch(output)

	if len(matches) < 4 {
		return 0
	}

	hours, _ := strconv.Atoi(matches[1])
	minutes, _ := strconv.Atoi(matches[2])
	seconds, _ := strconv.ParseFloat(matches[3], 64)

	// Calculate current time in seconds
	currentTime := float64(hours*3600+minutes*60) + seconds

	// Calculate percentage
	if duration > 0 {
		return int((currentTime / duration) * 100)
	}

	return 0
}

// getDuration gets the duration of a video file in seconds
func getDuration(filePath string) (float64, error) {
	cmd := exec.Command("ffmpeg", "-i", filePath)
	output, _ := cmd.CombinedOutput()

	// Find duration in output
	re := regexp.MustCompile(`Duration: (\d+):(\d+):(\d+\.\d+)`)
	matches := re.FindStringSubmatch(string(output))

	if len(matches) < 4 {
		return 0, fmt.Errorf("could not find duration")
	}

	hours, _ := strconv.Atoi(matches[1])
	minutes, _ := strconv.Atoi(matches[2])
	seconds, _ := strconv.ParseFloat(matches[3], 64)

	return float64(hours*3600+minutes*60) + seconds, nil
}

// transcodeVideo converts a video to streaming-friendly format
func transcodeVideo(c *fiber.Ctx) error {
	// Check if FFmpeg is installed
	if err := checkFFmpeg(); err != nil {
		log.Printf("FFmpeg error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("FFmpeg error: %v", err),
		})
	}

	// Get video ID from params
	id := c.Params("id")
	log.Printf("Transcoding request received for video: %s", id)

	// Get transcoding options from form
	format := c.FormValue("format", "mp4")
	resolution := c.FormValue("resolution", "720")
	bitrate := c.FormValue("bitrate", "1000k")

	// Validate format
	format = strings.ToLower(format)
	if format != "mp4" && format != "hls" && format != "dash" {
		format = "mp4" // Default to MP4
	}

	// Validate resolution
	validResolutions := map[string]bool{
		"240": true, "360": true, "480": true, "720": true,
		"1080": true, "1440": true, "2160": true,
	}
	if !validResolutions[resolution] {
		resolution = "720" // Default to 720p
	}

	// Validate bitrate
	validBitrates := map[string]bool{
		"500k": true, "1000k": true, "2000k": true,
		"4000k": true, "8000k": true, "16000k": true,
	}
	if !validBitrates[bitrate] {
		bitrate = "1000k" // Default to 1000k
	}

	// Create source and destination paths
	sourcePath := filepath.Join(uploadsDir, id)

	// Check if source file exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		log.Printf("Source video not found: %s", sourcePath)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Source video not found",
		})
	}

	// Get the duration of the video
	duration, err := getDuration(sourcePath)
	if err != nil {
		log.Printf("Failed to get video duration: %v", err)
		// Continue anyway, progress will be estimated
	}

	// Initialize progress for this video
	transcodingProgress[id] = 0

	// Create base name without extension
	baseName := strings.TrimSuffix(id, filepath.Ext(id))

	// Set output path based on format
	var outputPath string
	var outputUrl string

	switch format {
	case "hls":
		// For HLS, create a directory and output segments
		outputPath = filepath.Join(transcodedDir, baseName, "playlist.m3u8")
		outputDir := filepath.Join(transcodedDir, baseName)

		// Create directory
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			log.Printf("Failed to create transcoded directory for HLS: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to create transcoded directory: %v", err),
			})
		}

		// FFmpeg command for HLS with progress
		cmd := exec.Command("ffmpeg", "-i", sourcePath,
			"-profile:v", "baseline",
			"-level", "3.0",
			"-start_number", "0",
			"-hls_time", "10",
			"-hls_list_size", "0",
			"-f", "hls",
			"-vf", fmt.Sprintf("scale=-2:%s", resolution),
			"-b:v", bitrate,
			"-progress", "pipe:1", // Output progress to stdout
			outputPath)

		log.Printf("Running FFmpeg command: %v", cmd.String())

		// Run the command and capture output for progress
		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			log.Printf("Failed to create stdout pipe: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to create pipe: %v", err),
			})
		}

		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			log.Printf("Failed to create stderr pipe: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to create pipe: %v", err),
			})
		}

		// Start the command
		if err := cmd.Start(); err != nil {
			log.Printf("Failed to start FFmpeg: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to start FFmpeg: %v", err),
			})
		}

		// Read output and update progress in a goroutine
		go func() {
			buffer := make([]byte, 1024)
			for {
				n, err := stdoutPipe.Read(buffer)
				if n > 0 {
					output := string(buffer[:n])
					progress := parseProgress(output, duration)
					if progress > 0 && progress <= 100 {
						transcodingProgress[id] = progress
					}
				}
				if err != nil {
					break
				}
			}
		}()

		go func() {
			buffer := make([]byte, 1024)
			allOutput := ""
			for {
				n, err := stderrPipe.Read(buffer)
				if n > 0 {
					output := string(buffer[:n])
					allOutput += output
				}
				if err != nil {
					break
				}
			}
			// Store the full output for debugging
			log.Printf("FFmpeg stderr: %s", allOutput)
		}()

		// Wait for the command to finish
		if err := cmd.Wait(); err != nil {
			log.Printf("FFmpeg transcoding failed: %v", err)
			transcodingProgress[id] = -1 // -1 means error
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Transcoding failed: %v", err),
			})
		}

		// Set progress to 100% when done
		transcodingProgress[id] = 100
		outputUrl = fmt.Sprintf("/transcoded/%s/playlist.m3u8", baseName)

	case "dash":
		// For DASH, create a directory and output MPD file
		outputPath = filepath.Join(transcodedDir, baseName, "manifest.mpd")
		outputDir := filepath.Join(transcodedDir, baseName)

		// Create directory
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			log.Printf("Failed to create transcoded directory for DASH: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to create transcoded directory: %v", err),
			})
		}

		// FFmpeg command for DASH with progress
		cmd := exec.Command("ffmpeg", "-i", sourcePath,
			"-profile:v", "baseline",
			"-level", "3.0",
			"-bf", "0",
			"-f", "dash",
			"-vf", fmt.Sprintf("scale=-2:%s", resolution),
			"-b:v", bitrate,
			"-use_timeline", "1",
			"-use_template", "1",
			"-window_size", "5",
			"-adaptation_sets", "id=0,streams=v id=1,streams=a",
			"-progress", "pipe:1", // Output progress to stdout
			outputPath)

		log.Printf("Running FFmpeg command: %v", cmd.String())

		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			log.Printf("Failed to create stdout pipe: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to create pipe: %v", err),
			})
		}

		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			log.Printf("Failed to create stderr pipe: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to create pipe: %v", err),
			})
		}

		if err := cmd.Start(); err != nil {
			log.Printf("Failed to start FFmpeg: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to start FFmpeg: %v", err),
			})
		}

		go func() {
			buffer := make([]byte, 1024)
			for {
				n, err := stdoutPipe.Read(buffer)
				if n > 0 {
					output := string(buffer[:n])
					progress := parseProgress(output, duration)
					if progress > 0 && progress <= 100 {
						transcodingProgress[id] = progress
					}
				}
				if err != nil {
					break
				}
			}
		}()

		go func() {
			buffer := make([]byte, 1024)
			allOutput := ""
			for {
				n, err := stderrPipe.Read(buffer)
				if n > 0 {
					output := string(buffer[:n])
					allOutput += output
				}
				if err != nil {
					break
				}
			}
			log.Printf("FFmpeg stderr: %s", allOutput)
		}()

		if err := cmd.Wait(); err != nil {
			log.Printf("FFmpeg transcoding failed: %v", err)
			transcodingProgress[id] = -1
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Transcoding failed: %v", err),
			})
		}

		transcodingProgress[id] = 100
		outputUrl = fmt.Sprintf("/transcoded/%s/manifest.mpd", baseName)

	default: // mp4
		// For MP4, just output to a file
		outputPath = filepath.Join(transcodedDir, fmt.Sprintf("%s_%sp.mp4", baseName, resolution))

		// FFmpeg command for MP4 with progress
		cmd := exec.Command("ffmpeg", "-i", sourcePath,
			"-c:v", "libx264",
			"-preset", "fast",
			"-c:a", "aac",
			"-vf", fmt.Sprintf("scale=-2:%s", resolution),
			"-b:v", bitrate,
			"-movflags", "+faststart",
			"-progress", "pipe:1", // Output progress to stdout
			outputPath)

		log.Printf("Running FFmpeg command: %v", cmd.String())

		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			log.Printf("Failed to create stdout pipe: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to create pipe: %v", err),
			})
		}

		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			log.Printf("Failed to create stderr pipe: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to create pipe: %v", err),
			})
		}

		if err := cmd.Start(); err != nil {
			log.Printf("Failed to start FFmpeg: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to start FFmpeg: %v", err),
			})
		}

		go func() {
			buffer := make([]byte, 1024)
			for {
				n, err := stdoutPipe.Read(buffer)
				if n > 0 {
					output := string(buffer[:n])
					progress := parseProgress(output, duration)
					if progress > 0 && progress <= 100 {
						transcodingProgress[id] = progress
					}
				}
				if err != nil {
					break
				}
			}
		}()

		go func() {
			buffer := make([]byte, 1024)
			allOutput := ""
			for {
				n, err := stderrPipe.Read(buffer)
				if n > 0 {
					output := string(buffer[:n])
					allOutput += output
				}
				if err != nil {
					break
				}
			}
			log.Printf("FFmpeg stderr: %s", allOutput)
		}()

		if err := cmd.Wait(); err != nil {
			log.Printf("FFmpeg transcoding failed: %v", err)
			transcodingProgress[id] = -1
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Transcoding failed: %v", err),
			})
		}

		transcodingProgress[id] = 100
		outputUrl = fmt.Sprintf("/transcoded/%s_%sp.mp4", baseName, resolution)
	}

	// Return success response after transcoding is complete
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success":    true,
		"videoId":    id,
		"format":     format,
		"resolution": resolution,
		"url":        outputUrl,
	})
}

// getVideos returns a list of all videos
func getVideos(c *fiber.Ctx) error {
	// Read videos directory
	files, err := os.ReadDir(uploadsDir)
	if err != nil {
		log.Printf("Failed to read videos directory: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read videos directory",
		})
	}

	var videos []fiber.Map
	for _, file := range files {
		if !file.IsDir() {
			ext := filepath.Ext(file.Name())
			if ext == ".mp4" || ext == ".webm" || ext == ".mov" {
				videoName := file.Name()
				videoId := file.Name()

				// Check if it has transcoded versions
				baseName := strings.TrimSuffix(videoId, filepath.Ext(videoId))
				hasHLS := false
				hasDASH := false
				hasMP4 := false

				// Check for HLS version
				hlsPath := filepath.Join(transcodedDir, baseName, "playlist.m3u8")
				if _, err := os.Stat(hlsPath); err == nil {
					hasHLS = true
				}

				// Check for DASH version
				dashPath := filepath.Join(transcodedDir, baseName, "manifest.mpd")
				if _, err := os.Stat(dashPath); err == nil {
					hasDASH = true
				}

				// Look for MP4 versions
				mp4Files, _ := filepath.Glob(filepath.Join(transcodedDir, fmt.Sprintf("%s_*p.mp4", baseName)))
				hasMP4 = len(mp4Files) > 0

				videos = append(videos, fiber.Map{
					"id":      videoId,
					"name":    videoName,
					"url":     "/videos/" + videoId,
					"hasHLS":  hasHLS,
					"hasDASH": hasDASH,
					"hasMP4":  hasMP4,
					"hlsUrl": func() string {
						if hasHLS {
							return fmt.Sprintf("/transcoded/%s/playlist.m3u8", baseName)
						}
						return ""
					}(),
					"dashUrl": func() string {
						if hasDASH {
							return fmt.Sprintf("/transcoded/%s/manifest.mpd", baseName)
						}
						return ""
					}(),
				})
			}
		}
	}

	log.Printf("Returning %d videos", len(videos))
	return c.JSON(videos)
}

// getVideo returns a specific video by ID
func getVideo(c *fiber.Ctx) error {
	id := c.Params("id")
	videoPath := filepath.Join(uploadsDir, id)

	// Check if file exists
	_, err := os.Stat(videoPath)
	if os.IsNotExist(err) {
		log.Printf("Video not found: %s", id)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Video not found",
		})
	}

	// Create base name without extension
	baseName := strings.TrimSuffix(id, filepath.Ext(id))

	// Check for transcoded versions
	hasHLS := false
	hasDASH := false
	var mp4Versions []string

	// Check for HLS version
	hlsPath := filepath.Join(transcodedDir, baseName, "playlist.m3u8")
	if _, err := os.Stat(hlsPath); err == nil {
		hasHLS = true
	}

	// Check for DASH version
	dashPath := filepath.Join(transcodedDir, baseName, "manifest.mpd")
	if _, err := os.Stat(dashPath); err == nil {
		hasDASH = true
	}

	// Look for MP4 versions
	mp4Files, _ := filepath.Glob(filepath.Join(transcodedDir, fmt.Sprintf("%s_*p.mp4", baseName)))
	for _, file := range mp4Files {
		fileName := filepath.Base(file)
		mp4Versions = append(mp4Versions, "/transcoded/"+fileName)
	}

	// Return video info
	log.Printf("Returning video info: %s", id)
	return c.JSON(fiber.Map{
		"id":      id,
		"name":    id,
		"url":     "/videos/" + id,
		"hasHLS":  hasHLS,
		"hasDASH": hasDASH,
		"hlsUrl": func() string {
			if hasHLS {
				return fmt.Sprintf("/transcoded/%s/playlist.m3u8", baseName)
			}
			return ""
		}(),
		"dashUrl": func() string {
			if hasDASH {
				return fmt.Sprintf("/transcoded/%s/manifest.mpd", baseName)
			}
			return ""
		}(),
		"mp4Versions": mp4Versions,
	})
}

// uploadVideo handles video file uploads
func uploadVideo(c *fiber.Ctx) error {
	// Get file from request
	file, err := c.FormFile("video")
	if err != nil {
		log.Printf("No video file provided or error parsing form: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("No video file provided or error parsing form: %v", err),
		})
	}

	// Validate file size
	maxSize := 2000 * 1024 * 1024 // 2GB
	if file.Size > int64(maxSize) {
		log.Printf("File too large: %d bytes (max %d bytes)", file.Size, maxSize)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("File too large: %d bytes (max %d bytes)", file.Size, maxSize),
		})
	}

	// Get filename and extension
	filename := file.Filename
	ext := filepath.Ext(filename)

	// If extension is empty or not supported, use default extension
	if ext == "" || (ext != ".mp4" && ext != ".webm" && ext != ".mov") {
		log.Printf("Using default extension for file: %s", filename)
		filename = filename + ".mp4"
		ext = ".mp4"
	}

	// Ensure directory exists
	if err := os.MkdirAll(uploadsDir, os.ModePerm); err != nil {
		log.Printf("Failed to create uploads directory: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create uploads directory: %v", err),
		})
	}

	// Save file
	savePath := filepath.Join(uploadsDir, filename)
	log.Printf("Saving video to: %s", savePath)

	if err := c.SaveFile(file, savePath); err != nil {
		log.Printf("Failed to save video: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to save video: %v", err),
		})
	}

	// Verify file was saved
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		log.Printf("File was not saved properly: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "File was not saved properly after upload",
		})
	}

	log.Printf("Video uploaded successfully: %s (%d bytes)", filename, file.Size)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"id":   filename,
		"name": filename,
		"url":  "/videos/" + filename,
		"size": file.Size,
	})
}

// deleteVideo deletes a video by ID
func deleteVideo(c *fiber.Ctx) error {
	id := c.Params("id")
	log.Printf("Delete request received for video: %s", id)

	videoPath := filepath.Join(uploadsDir, id)
	log.Printf("Video path: %s", videoPath)

	// Check if file exists
	_, err := os.Stat(videoPath)
	if os.IsNotExist(err) {
		log.Printf("Video not found at path: %s", videoPath)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Video not found",
		})
	}

	// Delete file
	if err := os.Remove(videoPath); err != nil {
		log.Printf("Failed to delete video: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to delete video: %v", err),
		})
	}

	// Also delete any transcoded versions
	baseName := strings.TrimSuffix(id, filepath.Ext(id))

	// Delete HLS directory if exists
	hlsDir := filepath.Join(transcodedDir, baseName)
	os.RemoveAll(hlsDir) // Ignore errors

	// Delete any MP4 versions
	mp4Files, _ := filepath.Glob(filepath.Join(transcodedDir, fmt.Sprintf("%s_*p.mp4", baseName)))
	for _, file := range mp4Files {
		os.Remove(file) // Ignore errors
	}

	log.Printf("Video and transcoded versions successfully deleted: %s", id)
	return c.SendStatus(fiber.StatusNoContent)
}
