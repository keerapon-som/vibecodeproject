"use client";

import { useState, useEffect, useRef } from "react";
import Hls from "hls.js";

interface Video {
  id: string;
  name: string;
  url: string;
  hasHLS?: boolean;
  hasDASH?: boolean;
  hasMP4?: boolean;
  hlsUrl?: string;
  dashUrl?: string;
  mp4Versions?: string[];
}

interface TranscodeOptions {
  format: "mp4" | "hls" | "dash";
  resolution: "240" | "360" | "480" | "720" | "1080" | "1440" | "2160";
  bitrate: "500k" | "1000k" | "2000k" | "4000k" | "8000k" | "16000k";
}

export default function Home() {
  const [videos, setVideos] = useState<Video[]>([]);
  const [selectedVideo, setSelectedVideo] = useState<Video | null>(null);
  const [loading, setLoading] = useState(true);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [uploading, setUploading] = useState(false);
  const [autoReplay, setAutoReplay] = useState(false);
  const [transcoding, setTranscoding] = useState(false);
  const [transcodingProgress, setTranscodingProgress] = useState(0);
  const [showTranscodeOptions, setShowTranscodeOptions] = useState(false);
  const [transcodeOptions, setTranscodeOptions] = useState<TranscodeOptions>({
    format: "mp4",
    resolution: "1080",
    bitrate: "4000k"
  });
  const videoRef = useRef<HTMLVideoElement>(null);
  const API_URL = "http://localhost:8080";

  useEffect(() => {
    fetchVideos();
  }, []);

  // Set up the auto replay event listener
  useEffect(() => {
    const videoElement = videoRef.current;
    
    const handleVideoEnded = () => {
      if (autoReplay && videoElement) {
        videoElement.currentTime = 0;
        videoElement.play();
      }
    };

    if (videoElement) {
      videoElement.addEventListener('ended', handleVideoEnded);
    }

    // Clean up
    return () => {
      if (videoElement) {
        videoElement.removeEventListener('ended', handleVideoEnded);
      }
    };
  }, [autoReplay, selectedVideo]);

  // Progress polling for transcoding
  useEffect(() => {
    let interval: NodeJS.Timeout | null = null;

    if (transcoding && selectedVideo) {
      interval = setInterval(async () => {
        try {
          const encodedId = encodeURIComponent(selectedVideo.id);
          const response = await fetch(`${API_URL}/api/transcode/progress/${encodedId}`);
          if (response.ok) {
            const data = await response.json();
            if (data.progress !== undefined) {
              setTranscodingProgress(data.progress);

              // Stop polling when complete
              if (data.progress >= 100) {
                if (interval) clearInterval(interval);
              }
            }
          }
        } catch (error) {
          console.error("Error fetching transcoding progress:", error);
        }
      }, 1000);
    }

    return () => {
      if (interval) clearInterval(interval);
    };
  }, [transcoding, selectedVideo, API_URL]);

  const fetchVideos = async () => {
    try {
      setLoading(true);
      const response = await fetch(`${API_URL}/api/videos`);
      if (!response.ok) {
        throw new Error("Failed to fetch videos");
      }
      const data = await response.json();
      setVideos(data || []);
      setLoading(false);
    } catch (error) {
      console.error("Error fetching videos:", error);
      setLoading(false);
      setVideos([]);
    }
  };

  const handleVideoUpload = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const files = event.target.files;
    if (!files || files.length === 0) return;

    const file = files[0];
    const formData = new FormData();
    formData.append("video", file);

    try {
      setUploading(true);
      setUploadProgress(0);
      console.log(`Attempting to upload file: ${file.name} (${file.size} bytes)`);

      // Using fetch API instead of XMLHttpRequest
      const response = await fetch(`${API_URL}/api/videos`, {
        method: 'POST',
        body: formData,
      });

      console.log(`Upload response status: ${response.status}`);

      if (!response.ok) {
        let errorMessage = 'Upload failed';
        try {
          const errorData = await response.json();
          errorMessage = errorData.error || errorMessage;
        } catch (e) {
          // If we can't parse the response, just use the default error message
        }
        console.error(`Upload failed with status ${response.status}: ${errorMessage}`);
        alert(`Upload failed: ${errorMessage}`);
        setUploading(false);
        return;
      }

      console.log('Upload successful');
      setUploading(false);
      setUploadProgress(100);
      fetchVideos();
    } catch (error) {
      console.error("Error uploading video:", error);
      alert(`Error during upload: ${error instanceof Error ? error.message : 'Network error occurred'}`);
      setUploading(false);
    }
  };

  const handleDeleteVideo = async (videoId: string) => {
    try {
      // Make sure videoId is properly encoded
      console.log(`Attempting to delete video with ID: ${videoId}`);
      const encodedId = encodeURIComponent(videoId);
      console.log(`Encoded ID: ${encodedId}`);
      
      const deleteUrl = `${API_URL}/api/videos/${encodedId}`;
      console.log(`Delete URL: ${deleteUrl}`);
      
      const response = await fetch(deleteUrl, {
        method: "DELETE",
      });
      
      console.log(`Delete response status: ${response.status}`);
      
      if (!response.ok) {
        let errorText = "";
        try {
          const errorData = await response.json();
          errorText = errorData.error || "Unknown error";
        } catch (e) {
          errorText = "Could not parse error response";
        }
        console.error(`Error details: ${errorText}`);
        throw new Error(`Failed to delete video: ${errorText}`);
      }
      
      // Remove from selected if it was the active video
      if (selectedVideo && selectedVideo.id === videoId) {
        setSelectedVideo(null);
      }
      
      // Refresh the video list
      fetchVideos();
    } catch (error) {
      console.error("Error deleting video:", error);
      alert("Failed to delete video. Check console for details.");
    }
  };

  const handleTranscodeVideo = async () => {
    if (!selectedVideo) return;

    try {
      setTranscoding(true);
      setTranscodingProgress(0);
      const formData = new FormData();
      formData.append("format", transcodeOptions.format);
      formData.append("resolution", transcodeOptions.resolution);
      formData.append("bitrate", transcodeOptions.bitrate);

      const encodedId = encodeURIComponent(selectedVideo.id);
      const response = await fetch(`${API_URL}/api/videos/transcode/${encodedId}`, {
        method: "POST",
        body: formData,
      });

      if (!response.ok) {
        let errorMessage = 'Transcoding failed';
        try {
          const errorData = await response.json();
          errorMessage = errorData.error || errorMessage;
        } catch (e) {
          // If we can't parse the response, just use the default error message
        }
        throw new Error(errorMessage);
      }

      const result = await response.json();
      console.log("Transcoding successful:", result);

      // Refresh videos to update transcoded versions
      await fetchVideos();
      
      // Refresh selected video details
      const refreshResponse = await fetch(`${API_URL}/api/videos/${encodedId}`);
      if (refreshResponse.ok) {
        const updatedVideo = await refreshResponse.json();
        setSelectedVideo(updatedVideo);
      }

      setShowTranscodeOptions(false);
      alert("Video transcoded successfully!");
    } catch (error) {
      console.error("Error transcoding video:", error);
      alert(`Transcoding failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    } finally {
      setTranscoding(false);
      setTranscodingProgress(0);
    }
  };

  const handleReplayVideo = () => {
    if (videoRef.current) {
      videoRef.current.currentTime = 0;
      videoRef.current.play();
    }
  };

  const toggleAutoReplay = () => {
    setAutoReplay(!autoReplay);
  };

  const getVideoSrc = () => {
    if (!selectedVideo) return "";

    // Use HLS if available and supported
    if (selectedVideo.hasHLS && selectedVideo.hlsUrl && Hls.isSupported()) {
      return selectedVideo.hlsUrl;
    } 
    // Use DASH if available and supported
    else if (selectedVideo.hasDASH && selectedVideo.dashUrl && 'MediaSource' in window) {
      return selectedVideo.dashUrl;
    }
    // Use transcoded MP4 if available
    else if (selectedVideo.hasMP4 && selectedVideo.mp4Versions && selectedVideo.mp4Versions.length > 0) {
      return selectedVideo.mp4Versions[0]; // Use the first MP4 version
    }
    // Use original video
    else {
      return `${API_URL}${selectedVideo.url}`;
    }
  };

  // Initialize HLS.js if needed
  useEffect(() => {
    // Check if using HLS
    if (selectedVideo?.hasHLS && selectedVideo.hlsUrl && videoRef.current) {
      // Check if HLS.js is supported
      if (typeof Hls !== 'undefined' && Hls.isSupported()) {
        const hls = new Hls();
        hls.loadSource(`${API_URL}${selectedVideo.hlsUrl}`);
        hls.attachMedia(videoRef.current);
        
        return () => {
          hls.destroy();
        };
      }
    }
  }, [selectedVideo]);

  return (
    <main className="flex min-h-screen flex-col p-8">
      <h1 className="text-4xl font-bold mb-8 text-center">Video Streaming Platform</h1>
      
      <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
        {/* Video Player */}
        <div className="md:col-span-2 bg-gray-900 rounded-lg overflow-hidden">
          {selectedVideo ? (
            <div>
              <video 
                ref={videoRef}
                src={getVideoSrc()}
                controls
                autoPlay
                className="w-full h-auto"
              />
              <div className="p-4">
                <h2 className="text-xl font-semibold mb-2">{selectedVideo.name}</h2>
                <div className="flex flex-wrap items-center gap-4 mb-2">
                  <div className="flex items-center">
                    <label className="inline-flex items-center cursor-pointer">
                      <input 
                        type="checkbox" 
                        checked={autoReplay} 
                        onChange={toggleAutoReplay} 
                        className="sr-only peer"
                      />
                      <div className="relative w-11 h-6 bg-gray-700 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
                      <span className="ms-3 text-sm font-medium text-white">Auto Replay</span>
                    </label>
                  </div>

                  <button
                    onClick={() => setShowTranscodeOptions(!showTranscodeOptions)}
                    className="px-4 py-2 bg-purple-600 text-white rounded hover:bg-purple-700 transition"
                    disabled={transcoding}
                  >
                    {transcoding ? "Transcoding..." : "Transcode Video"}
                  </button>
                </div>

                {showTranscodeOptions && (
                  <div className="mt-4 p-4 bg-gray-800 rounded-lg">
                    <h3 className="text-lg font-medium mb-3">Transcoding Options</h3>
                    
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-300 mb-1">Format</label>
                        <select 
                          value={transcodeOptions.format}
                          onChange={(e) => setTranscodeOptions({...transcodeOptions, format: e.target.value as any})}
                          className="w-full bg-gray-700 text-white p-2 rounded"
                        >
                          <option value="mp4">MP4</option>
                          <option value="hls">HLS (Streaming)</option>
                          <option value="dash">DASH (Adaptive)</option>
                        </select>
                      </div>
                      
                      <div>
                        <label className="block text-sm font-medium text-gray-300 mb-1">Resolution</label>
                        <select 
                          value={transcodeOptions.resolution}
                          onChange={(e) => setTranscodeOptions({...transcodeOptions, resolution: e.target.value as any})}
                          className="w-full bg-gray-700 text-white p-2 rounded"
                        >
                          <option value="240">240p</option>
                          <option value="360">360p</option>
                          <option value="480">480p</option>
                          <option value="720">720p (HD)</option>
                          <option value="1080">1080p (Full HD)</option>
                          <option value="1440">1440p (2K)</option>
                          <option value="2160">2160p (4K)</option>
                        </select>
                      </div>
                      
                      <div>
                        <label className="block text-sm font-medium text-gray-300 mb-1">Bitrate</label>
                        <select 
                          value={transcodeOptions.bitrate}
                          onChange={(e) => setTranscodeOptions({...transcodeOptions, bitrate: e.target.value as any})}
                          className="w-full bg-gray-700 text-white p-2 rounded"
                        >
                          <option value="500k">Very Low (500k)</option>
                          <option value="1000k">Low (1000k)</option>
                          <option value="2000k">Medium (2000k)</option>
                          <option value="4000k">High (4000k)</option>
                          <option value="8000k">Very High (8000k)</option>
                          <option value="16000k">Ultra (16000k)</option>
                        </select>
                      </div>
                    </div>
                    
                    {transcoding && (
                      <div className="mb-4">
                        <div className="flex justify-between text-sm text-gray-300 mb-1">
                          <span>Transcoding Progress</span>
                          <span>{transcodingProgress}%</span>
                        </div>
                        <div className="w-full bg-gray-700 rounded-full h-2.5">
                          <div
                            className="bg-green-600 h-2.5 rounded-full transition-all duration-300"
                            style={{ width: `${transcodingProgress}%` }}
                          ></div>
                        </div>
                      </div>
                    )}
                    
                    <div className="flex justify-end">
                      <button
                        onClick={handleTranscodeVideo}
                        disabled={transcoding}
                        className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 transition disabled:opacity-50"
                      >
                        {transcoding ? `Processing... (${transcodingProgress}%)` : "Start Transcoding"}
                      </button>
                    </div>
                  </div>
                )}

                {/* Transcoded versions */}
                {selectedVideo.hasHLS || selectedVideo.hasDASH || (selectedVideo.mp4Versions && selectedVideo.mp4Versions.length > 0) ? (
                  <div className="mt-4">
                    <h3 className="text-md font-medium mb-2">Available Formats:</h3>
                    <div className="flex flex-wrap gap-2">
                      {selectedVideo.hasHLS && (
                        <span className="px-2 py-1 bg-blue-600 text-white text-xs rounded">HLS</span>
                      )}
                      {selectedVideo.hasDASH && (
                        <span className="px-2 py-1 bg-green-600 text-white text-xs rounded">DASH</span>
                      )}
                      {selectedVideo.mp4Versions && selectedVideo.mp4Versions.map((version, index) => {
                        // Extract resolution from filename (e.g., video_720p.mp4)
                        const resolution = version.match(/_(\d+)p\.mp4$/)?.[1] || 'MP4';
                        return (
                          <span key={index} className="px-2 py-1 bg-purple-600 text-white text-xs rounded">
                            {resolution}p
                          </span>
                        );
                      })}
                    </div>
                  </div>
                ) : null}
                
                {autoReplay && (
                  <p className="text-sm text-green-400 mt-2">Auto replay is enabled. Video will automatically restart when finished.</p>
                )}
              </div>
            </div>
          ) : (
            <div className="flex items-center justify-center h-[400px]">
              <p className="text-gray-400">Select a video to play</p>
            </div>
          )}
        </div>
        
        {/* Video List and Upload */}
        <div className="bg-gray-800 rounded-lg p-4">
          <div className="mb-6">
            <h3 className="text-lg font-medium mb-3">Upload Video</h3>
            <div className="flex flex-col">
              <input
                type="file"
                accept="video/*"
                onChange={handleVideoUpload}
                className="mb-2"
                disabled={uploading}
              />
              {uploading && (
                <div className="w-full bg-gray-700 rounded-full h-2.5 mb-4">
                  <div
                    className="bg-blue-600 h-2.5 rounded-full"
                    style={{ width: `${uploadProgress}%` }}
                  ></div>
                </div>
              )}
            </div>
          </div>
          
          <h3 className="text-lg font-medium mb-3">Videos</h3>
          {loading ? (
            <p>Loading videos...</p>
          ) : videos && videos.length > 0 ? (
            <ul className="divide-y divide-gray-700">
              {videos.map((video) => (
                <li key={video.id} className="py-3">
                  <div className="flex justify-between items-center">
                    <button
                      onClick={() => setSelectedVideo(video)}
                      className="text-left hover:text-blue-400 transition truncate flex-1"
                    >
                      {video.name}
                      {(video.hasHLS || video.hasDASH || video.hasMP4) && (
                        <span className="ml-2 px-1.5 py-0.5 bg-green-700 text-white text-xs rounded">
                          Transcoded
                        </span>
                      )}
                    </button>
                    <button
                      onClick={() => handleDeleteVideo(video.id)}
                      className="text-red-500 hover:text-red-400 ml-2"
                    >
                      Delete
                    </button>
                  </div>
                </li>
              ))}
            </ul>
          ) : (
            <p className="text-gray-400">No videos available</p>
          )}
        </div>
      </div>
    </main>
  );
}
