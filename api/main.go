package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Job struct {
	ID             string    `json:"id"`
	Status         string    `json:"status"`
	URL            string    `json:"url,omitempty"` // Store URL for resumption
	File           string    `json:"file,omitempty"`
	AudioFile      string    `json:"audio_file,omitempty"`
	Text           string    `json:"text,omitempty"`
	Progress       int       `json:"progress"`
	AudioProgress  int       `json:"audio_progress"`
	TranscriptProgress int   `json:"transcript_progress"`
	Error          string    `json:"error,omitempty"`
	Created        time.Time `json:"created"`
	
	// Video metadata
	Title          string    `json:"title,omitempty"`
	Description    string    `json:"description,omitempty"`
	Thumbnail      string    `json:"thumbnail,omitempty"`
	Duration       int       `json:"duration,omitempty"` // Duration in seconds
	ChannelName    string    `json:"channel_name,omitempty"`
}

var (
	jobs   = make(map[string]*Job)
	jobsMu sync.RWMutex
	
	// Job queue for background processing
	jobQueue = make(chan string, 100) // Buffer for job IDs
)

func main() {
	// Ensure data directory exists
	if err := os.MkdirAll("/data", 0755); err != nil {
		log.Printf("Warning: Could not create /data directory: %v", err)
	}
	
	// Ensure jobs directory exists for persistence
	if err := os.MkdirAll("/data/jobs", 0755); err != nil {
		log.Printf("Warning: Could not create /data/jobs directory: %v", err)
	}
	
	// Load existing jobs from disk
	loadJobsFromDisk()

	// Start background worker for job processing
	go backgroundWorker()

	// Setup HTTP handlers
	http.HandleFunc("/job", handleJob)
	http.HandleFunc("/job/", handleGetJob)
	http.HandleFunc("/jobs/active", handleGetActiveJobs)
	http.HandleFunc("/jobs/history", handleGetJobHistory)
	http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir("/data"))))

	// CORS middleware for local development
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		http.NotFound(w, r)
	})

	log.Println("Server starting on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func handleJob(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if payload.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Validate YouTube URL
	if !isValidYouTubeURL(payload.URL) {
		http.Error(w, "Invalid YouTube URL", http.StatusBadRequest)
		return
	}

	id := uuid.NewString()
	job := &Job{
		ID:       id,
		Status:   "queued",
		URL:      payload.URL, // Store URL for resumption
		Progress: 0,
		Created:  time.Now(),
	}

	jobsMu.Lock()
	jobs[id] = job
	jobsMu.Unlock()

	// Queue job for background processing instead of immediate processing
	select {
	case jobQueue <- id:
		log.Printf("Job %s queued for background processing", id)
	default:
		log.Printf("Job queue full, processing job %s immediately", id)
		go processJob(job, payload.URL)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func handleGetJob(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	id := strings.TrimPrefix(r.URL.Path, "/job/")
	if id == "" {
		http.Error(w, "Job ID required", http.StatusBadRequest)
		return
	}

	jobsMu.RLock()
	job, exists := jobs[id]
	jobsMu.RUnlock()

	if !exists {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(job)
}

func handleGetActiveJobs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobsMu.RLock()
	activeJobs := make([]*Job, 0)
	for _, job := range jobs {
		// Return jobs that are active or recently completed (last 24 hours)
		if job.Status != "done" && job.Status != "error" {
			activeJobs = append(activeJobs, job)
		} else {
			// Check if completed recently
			if time.Since(job.Created).Hours() < 24 {
				activeJobs = append(activeJobs, job)
			}
		}
	}
	jobsMu.RUnlock()

	json.NewEncoder(w).Encode(activeJobs)
}

func handleGetJobHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobsMu.RLock()
	historyJobs := make([]*Job, 0)
	for _, job := range jobs {
		// Return only completed jobs sorted by creation date (newest first)
		if job.Status == "done" {
			historyJobs = append(historyJobs, job)
		}
	}
	jobsMu.RUnlock()
	
	// Sort by creation date (newest first)
	for i := 0; i < len(historyJobs)-1; i++ {
		for j := i + 1; j < len(historyJobs); j++ {
			if historyJobs[i].Created.Before(historyJobs[j].Created) {
				historyJobs[i], historyJobs[j] = historyJobs[j], historyJobs[i]
			}
		}
	}

	json.NewEncoder(w).Encode(historyJobs)
}

func processJob(job *Job, url string) {
	defer func() {
		if r := recover(); r != nil {
			updateJobStatusDetailed(job, "error", 0, 0, 0, fmt.Sprintf("Internal error: %v", r))
		}
	}()

	// Step 0: Extract video metadata
	updateJobStatusDetailed(job, "fetching_info", 10, 0, 0, "")
	if err := extractVideoMetadata(job, url); err != nil {
		log.Printf("Warning: Failed to extract video metadata: %v", err)
		// Continue processing even if metadata extraction fails
	}

	// Step 1: Download audio
	updateJobStatusDetailed(job, "downloading", 25, 0, 0, "")
	audioFile, err := downloadAudio(job.ID, url)
	if err != nil {
		updateJobStatusDetailed(job, "error", 0, 0, 0, fmt.Sprintf("Download failed: %v", err))
		return
	}
	
	// Copy audio to data directory for serving
	audioFilename := fmt.Sprintf("%s.wav", job.ID)
	audioPath := filepath.Join("/data", audioFilename)
	if err := copyFile(audioFile, audioPath); err != nil {
		log.Printf("Warning: Failed to copy audio file for serving: %v", err)
	} else {
		jobsMu.Lock()
		job.AudioFile = "/files/" + audioFilename
		jobsMu.Unlock()
		log.Printf("Audio file available for download: /files/%s", audioFilename)
	}
	
	// Audio download complete
	updateJobStatusDetailed(job, "transcribing", 50, 100, 0, "")

	// Step 2: Transcribe
	transcript, err := transcribeAudio(audioFile)
	if err != nil {
		updateJobStatusDetailed(job, "error", 0, 100, 0, fmt.Sprintf("Transcription failed: %v", err))
		return
	}

	// Step 3: Save result
	updateJobStatusDetailed(job, "saving", 90, 100, 90, "")
	filename := fmt.Sprintf("%s.txt", job.ID)
	filepath := filepath.Join("/data", filename)

	if err := os.WriteFile(filepath, []byte(transcript), 0644); err != nil {
		updateJobStatusDetailed(job, "error", 0, 100, 0, fmt.Sprintf("Save failed: %v", err))
		return
	}

	// Complete
	jobsMu.Lock()
	job.Status = "done"
	job.Progress = 100
	job.AudioProgress = 100
	job.TranscriptProgress = 100
	job.Text = transcript
	job.File = "/files/" + filename
	jobsMu.Unlock()
	
	// Save final job state
	saveJobToDisk(job)
}

// processJobResume resumes an interrupted job
func processJobResume(job *Job) {
	defer func() {
		if r := recover(); r != nil {
			updateJobStatusDetailed(job, "error", 0, 0, 0, fmt.Sprintf("Internal error during resume: %v", r))
		}
	}()
	
	log.Printf("Resuming job %s from status: %s", job.ID, job.Status)
	
	// Check if audio file already exists
	audioFilename := fmt.Sprintf("%s.wav", job.ID)
	audioPath := filepath.Join("/data", audioFilename)
	tmpAudioPath := filepath.Join("/tmp", job.ID+".wav")
	
	var audioFile string
	
	// If audio already downloaded, use existing file
	if _, err := os.Stat(audioPath); err == nil {
		log.Printf("Found existing audio file for job %s", job.ID)
		audioFile = audioPath
		updateJobStatusDetailed(job, "transcribing", 50, 100, 0, "")
	} else if _, err := os.Stat(tmpAudioPath); err == nil {
		log.Printf("Found temporary audio file for job %s", job.ID)
		audioFile = tmpAudioPath
		// Copy to data directory
		if err := copyFile(audioFile, audioPath); err == nil {
			jobsMu.Lock()
			job.AudioFile = "/files/" + audioFilename
			jobsMu.Unlock()
		}
		updateJobStatusDetailed(job, "transcribing", 50, 100, 0, "")
	} else {
		// Need to re-download audio (job was very early when interrupted)
		log.Printf("No audio file found, need to restart download for job %s", job.ID)
		if job.URL == "" {
			updateJobStatusDetailed(job, "error", 0, 0, 0, "Cannot resume: original URL not saved")
			return
		}
		
		updateJobStatusDetailed(job, "downloading", 25, 0, 0, "")
		downloadedAudio, err := downloadAudio(job.ID, job.URL)
		if err != nil {
			updateJobStatusDetailed(job, "error", 0, 0, 0, fmt.Sprintf("Download failed: %v", err))
			return
		}
		
		// Copy audio to data directory for serving
		if err := copyFile(downloadedAudio, audioPath); err == nil {
			jobsMu.Lock()
			job.AudioFile = "/files/" + audioFilename
			jobsMu.Unlock()
		}
		
		audioFile = downloadedAudio
		updateJobStatusDetailed(job, "transcribing", 50, 100, 0, "")
	}
	
	// Check if transcript already exists
	filename := fmt.Sprintf("%s.txt", job.ID)
	filepath := filepath.Join("/data", filename)
	if _, err := os.Stat(filepath); err == nil {
		// Transcript already exists, load it
		if data, err := os.ReadFile(filepath); err == nil {
			jobsMu.Lock()
			job.Status = "done"
			job.Progress = 100
			job.AudioProgress = 100
			job.TranscriptProgress = 100
			job.Text = string(data)
			job.File = "/files/" + filename
			jobsMu.Unlock()
			saveJobToDisk(job)
			log.Printf("Job %s already completed, loaded existing transcript", job.ID)
			return
		}
	}
	
	// Continue with transcription
	transcript, err := transcribeAudio(audioFile)
	if err != nil {
		updateJobStatusDetailed(job, "error", 0, 100, 0, fmt.Sprintf("Transcription failed: %v", err))
		return
	}

	// Save result
	updateJobStatusDetailed(job, "saving", 90, 100, 90, "")
	if err := os.WriteFile(filepath, []byte(transcript), 0644); err != nil {
		updateJobStatusDetailed(job, "error", 0, 100, 0, fmt.Sprintf("Save failed: %v", err))
		return
	}

	// Complete
	jobsMu.Lock()
	job.Status = "done"
	job.Progress = 100
	job.AudioProgress = 100
	job.TranscriptProgress = 100
	job.Text = transcript
	job.File = "/files/" + filename
	jobsMu.Unlock()
	
	saveJobToDisk(job)
	log.Printf("Successfully resumed and completed job %s", job.ID)
}

func updateJobStatus(job *Job, status string, progress int, error string) {
	jobsMu.Lock()
	job.Status = status
	job.Progress = progress
	if error != "" {
		job.Error = error
	}
	jobsMu.Unlock()
}

func updateJobStatusDetailed(job *Job, status string, overallProgress, audioProgress, transcriptProgress int, error string) {
	jobsMu.Lock()
	job.Status = status
	job.Progress = overallProgress
	job.AudioProgress = audioProgress
	job.TranscriptProgress = transcriptProgress
	if error != "" {
		job.Error = error
	}
	jobsMu.Unlock()
	
	// Save job state to disk for persistence
	saveJobToDisk(job)
}

func copyFile(src, dst string) error {
	// Simple file copy using os.ReadFile/WriteFile
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file %s: %v", src, err)
	}
	
	err = os.WriteFile(dst, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write destination file %s: %v", dst, err)
	}
	
	log.Printf("Successfully copied audio file: %s -> %s (%d bytes)", src, dst, len(data))
	return nil
}

func extractVideoMetadata(job *Job, url string) error {
	// Use yt-dlp to get video metadata
	cmd := exec.Command("yt-dlp", "--dump-json", "--no-download", url)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to extract metadata: %v", err)
	}
	
	// Parse JSON output
	var metadata struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Thumbnail   string  `json:"thumbnail"`
		Duration    float64 `json:"duration"`
		Channel     string  `json:"channel"`
		Uploader    string  `json:"uploader"`
	}
	
	if err := json.Unmarshal(output, &metadata); err != nil {
		return fmt.Errorf("failed to parse metadata JSON: %v", err)
	}
	
	// Update job with metadata
	jobsMu.Lock()
	job.Title = metadata.Title
	job.Description = metadata.Description
	job.Thumbnail = metadata.Thumbnail
	job.Duration = int(metadata.Duration)
	job.ChannelName = metadata.Channel
	if job.ChannelName == "" {
		job.ChannelName = metadata.Uploader
	}
	jobsMu.Unlock()
	
	log.Printf("Extracted metadata for job %s: title=%s, duration=%ds, channel=%s", 
		job.ID, job.Title, job.Duration, job.ChannelName)
	
	return nil
}

func downloadAudio(jobID, url string) (string, error) {
	tmpFile := filepath.Join("/tmp", jobID+".wav")

	// Use yt-dlp to download best audio and pipe to ffmpeg for conversion
	ytCmd := exec.Command("yt-dlp",
		"-f", "ba",
		"-o", "-",
		url,
		"--quiet")

	ffCmd := exec.Command("ffmpeg",
		"-i", "pipe:0",
		"-vn",
		"-ac", "1",
		"-ar", "16000",
		"-f", "wav",
		tmpFile,
		"-y") // Overwrite output file

	// Connect yt-dlp output to ffmpeg input
	pipe, err := ytCmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create pipe: %v", err)
	}
	ffCmd.Stdin = pipe

	// Start both commands
	if err := ffCmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start ffmpeg: %v", err)
	}

	if err := ytCmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start yt-dlp: %v", err)
	}

	// Wait for completion
	if err := ytCmd.Wait(); err != nil {
		return "", fmt.Errorf("yt-dlp failed: %v", err)
	}

	if err := ffCmd.Wait(); err != nil {
		return "", fmt.Errorf("ffmpeg failed: %v", err)
	}

	return tmpFile, nil
}

func transcribeAudio(audioFile string) (string, error) {
	// Check audio file size and split if too large
	fileInfo, err := os.Stat(audioFile)
	if err != nil {
		return "", fmt.Errorf("cannot check audio file: %v", err)
	}
	
	// If file is larger than 10MB, split into chunks
	if fileInfo.Size() > 10*1024*1024 {
		log.Printf("Audio file is large (%d bytes), splitting into chunks", fileInfo.Size())
		return transcribeAudioChunked(audioFile)
	}
	
	// Process small files directly
	return transcribeAudioDirect(audioFile)
}

func transcribeAudioChunked(audioFile string) (string, error) {
	chunks, err := splitAudioFile(audioFile)
	if err != nil {
		return "", fmt.Errorf("failed to split audio: %v", err)
	}
	
	var fullTranscript strings.Builder
	for i, chunk := range chunks {
		log.Printf("Processing chunk %d/%d", i+1, len(chunks))
		
		text, err := transcribeAudioDirect(chunk)
		if err != nil {
			log.Printf("Chunk %d failed: %v", i+1, err)
			continue // Skip failed chunks rather than fail entirely
		}
		
		fullTranscript.WriteString(text)
		fullTranscript.WriteString(" ")
		
		// Clean up chunk file
		os.Remove(chunk)
	}
	
	return strings.TrimSpace(fullTranscript.String()), nil
}

func splitAudioFile(audioFile string) ([]string, error) {
	baseDir := filepath.Dir(audioFile)
	baseName := strings.TrimSuffix(filepath.Base(audioFile), ".wav")
	
	chunks := []string{}
	chunkDuration := 120 // 2 minutes per chunk
	
	// Use ffmpeg to split audio into 2-minute chunks
	for i := 0; i < 10; i++ { // Max 10 chunks (20 minutes total)
		chunkFile := filepath.Join(baseDir, fmt.Sprintf("%s_chunk_%d.wav", baseName, i))
		startTime := i * chunkDuration
		
		cmd := exec.Command("ffmpeg",
			"-i", audioFile,
			"-ss", fmt.Sprintf("%d", startTime),
			"-t", fmt.Sprintf("%d", chunkDuration),
			"-c", "copy",
			chunkFile,
			"-y") // Overwrite output
		
		err := cmd.Run()
		if err != nil {
			// No more audio to split
			break
		}
		
		// Check if chunk has content
		if info, err := os.Stat(chunkFile); err == nil && info.Size() > 1000 {
			chunks = append(chunks, chunkFile)
		} else {
			os.Remove(chunkFile)
			break
		}
	}
	
	return chunks, nil
}

func transcribeAudioDirect(audioFile string) (string, error) {
	log.Printf("Transcribing audio file: %s", filepath.Base(audioFile))
	
	// Use OpenAI Whisper binary for transcription
	baseName := strings.TrimSuffix(filepath.Base(audioFile), ".wav")
	outputDir := "/tmp"
	
	// Run whisper command
	cmd := exec.Command("whisper", 
		audioFile,
		"--model", "tiny",
		"--output_format", "txt",
		"--output_dir", outputDir,
		"--verbose", "False")
	
	// Capture both stdout and stderr for debugging
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Whisper command failed: %v, output: %s", err, string(output))
		return "", fmt.Errorf("whisper transcription failed: %v", err)
	}
	
	log.Printf("Whisper command completed successfully")
	
	// Read the generated transcript file
	transcriptFile := filepath.Join(outputDir, baseName+".txt")
	transcriptBytes, err := os.ReadFile(transcriptFile)
	if err != nil {
		log.Printf("Failed to read transcript file %s: %v", transcriptFile, err)
		return "", fmt.Errorf("failed to read transcript file: %v", err)
	}
	
	transcript := strings.TrimSpace(string(transcriptBytes))
	if transcript == "" {
		return "", fmt.Errorf("empty transcript generated")
	}
	
	// Clean up the transcript file
	os.Remove(transcriptFile)
	
	log.Printf("Successfully transcribed audio, transcript length: %d characters", len(transcript))
	return transcript, nil
}

func parseWhisperResponse(output []byte) (string, error) {
	// Parse JSON response (OpenAI Whisper API format)
	var response struct {
		Text string `json:"text"`
	}

	// Log the raw response for debugging
	log.Printf("Whisper response: %s", string(output))

	if err := json.Unmarshal(output, &response); err != nil {
		log.Printf("JSON parsing failed: %v, raw response: %s", err, string(output))
		// If JSON parsing fails, assume raw text response
		return string(output), nil
	}

	if response.Text == "" {
		return "", fmt.Errorf("empty text response from whisper")
	}

	// Validate transcript quality  
	if len(response.Text) < 10 {
		log.Printf("Warning: Very short transcript (%d chars): %s", len(response.Text), response.Text)
	}

	return response.Text, nil
}

func parseWhisperCppResponse(output []byte) string {
	// whisper.cpp returns different format than go-whisper
	response := string(output)
	log.Printf("Whisper.cpp raw response: %s", response)
	
	// Try to parse as JSON first
	var jsonResp struct {
		Text string `json:"text"`
	}
	
	if err := json.Unmarshal(output, &jsonResp); err == nil && jsonResp.Text != "" {
		return strings.TrimSpace(jsonResp.Text)
	}
	
	// If not JSON, treat as plain text response
	return strings.TrimSpace(response)
}

func isValidYouTubeURL(url string) bool {
	return strings.Contains(url, "youtube.com") ||
		strings.Contains(url, "youtu.be") ||
		strings.Contains(url, "m.youtube.com")
}

// saveJobToDisk saves job state to disk for persistence
func saveJobToDisk(job *Job) {
	jobFile := filepath.Join("/data/jobs", job.ID+".json")
	data, err := json.MarshalIndent(job, "", "  ")
	if err != nil {
		log.Printf("Error marshaling job %s: %v", job.ID, err)
		return
	}
	
	if err := os.WriteFile(jobFile, data, 0644); err != nil {
		log.Printf("Error saving job %s to disk: %v", job.ID, err)
	}
}

// backgroundWorker processes jobs from the queue one at a time
func backgroundWorker() {
	log.Println("Background worker started")
	
	for jobID := range jobQueue {
		log.Printf("Processing job %s from queue", jobID)
		
		jobsMu.RLock()
		job, exists := jobs[jobID]
		jobsMu.RUnlock()
		
		if !exists {
			log.Printf("Job %s not found in memory", jobID)
			continue
		}
		
		// Update job to processing status
		jobsMu.Lock()
		job.Status = "processing"
		jobsMu.Unlock()
		
		// Process the job
		processJob(job, job.URL)
		
		log.Printf("Completed processing job %s", jobID)
	}
}

// loadJobsFromDisk loads existing jobs from disk on startup
func loadJobsFromDisk() {
	jobsDir := "/data/jobs"
	files, err := os.ReadDir(jobsDir)
	if err != nil {
		log.Printf("Could not read jobs directory: %v", err)
		return
	}
	
	loadedCount := 0
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		
		jobFile := filepath.Join(jobsDir, file.Name())
		data, err := os.ReadFile(jobFile)
		if err != nil {
			log.Printf("Error reading job file %s: %v", jobFile, err)
			continue
		}
		
		var job Job
		if err := json.Unmarshal(data, &job); err != nil {
			log.Printf("Error unmarshaling job file %s: %v", jobFile, err)
			continue
		}
		
		jobsMu.Lock()
		jobs[job.ID] = &job
		jobsMu.Unlock()
		
		// Resume processing if job was interrupted
		if job.Status != "done" && job.Status != "error" {
			log.Printf("Resuming interrupted job: %s", job.ID)
			// Reset status to allow resumption
			job.Status = "queued"
			go processJobResume(&job)
		}
		
		loadedCount++
	}
	
	if loadedCount > 0 {
		log.Printf("Loaded %d jobs from disk", loadedCount)
	}
}
