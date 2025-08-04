package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestIsValidYouTubeURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ", true},
		{"https://youtu.be/dQw4w9WgXcQ", true},
		{"https://m.youtube.com/watch?v=dQw4w9WgXcQ", true},
		{"https://www.youtube.com/embed/dQw4w9WgXcQ", true},
		{"https://example.com", false},
		{"", false},
		{"not-a-url", false},
	}

	for _, test := range tests {
		result := isValidYouTubeURL(test.url)
		if result != test.expected {
			t.Errorf("isValidYouTubeURL(%q) = %v; want %v", test.url, result, test.expected)
		}
	}
}

func TestJobCreation(t *testing.T) {
	job := &Job{
		ID:       "test-id",
		Status:   "queued",
		Progress: 0,
	}

	if job.ID != "test-id" {
		t.Errorf("Expected job ID to be 'test-id', got %s", job.ID)
	}

	if job.Status != "queued" {
		t.Errorf("Expected job status to be 'queued', got %s", job.Status)
	}

	if job.Progress != 0 {
		t.Errorf("Expected job progress to be 0, got %d", job.Progress)
	}
}

// TestYouTubeTranscription tests the transcription of the specific YouTube video
func TestYouTubeTranscription(t *testing.T) {
	// Test URL: https://www.youtube.com/watch?v=c-P5R0aMylM
	testURL := "https://www.youtube.com/watch?v=c-P5R0aMylM"
	
	// Create test payload
	payload := `{"url":"` + testURL + `"}`
	
	// Create request
	req, err := http.NewRequest("POST", "/job", strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	// Create response recorder
	rr := httptest.NewRecorder()
	
	// Call handler
	handleJob(rr, req)
	
	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	// Parse response
	var job Job
	if err := json.Unmarshal(rr.Body.Bytes(), &job); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	
	// Verify job was created
	if job.ID == "" {
		t.Error("Job ID is empty")
	}
	
	if job.Status != "queued" {
		t.Errorf("Expected status 'queued', got '%s'", job.Status)
	}
	
	// Wait for processing to complete or timeout
	timeout := time.After(10 * time.Minute) // Long timeout for transcription
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-timeout:
			t.Fatal("Test timed out after 10 minutes")
		case <-ticker.C:
			// Check job status
			jobsMu.RLock()
			currentJob, exists := jobs[job.ID]
			jobsMu.RUnlock()
			
			if !exists {
				t.Fatal("Job not found")
			}
			
			t.Logf("Job status: %s, Progress: %d%%, Audio: %d%%, Transcript: %d%%", 
				currentJob.Status, currentJob.Progress, currentJob.AudioProgress, currentJob.TranscriptProgress)
			
			if currentJob.Status == "error" {
				t.Fatalf("Job failed with error: %s", currentJob.Error)
			}
			
			if currentJob.Status == "done" {
				// Verify transcript was generated
				if currentJob.Text == "" {
					t.Error("Transcript text is empty")
				}
				
				if currentJob.File == "" {
					t.Error("File path is empty")
				}
				
				if currentJob.AudioFile == "" {
					t.Error("Audio file path is empty")
				}
				
				// Log successful transcription details
				t.Logf("Successfully transcribed video. Transcript length: %d characters", len(currentJob.Text))
				if len(currentJob.Text) > 200 {
					t.Logf("First 200 characters: %s...", currentJob.Text[:200])
				} else {
					t.Logf("Full transcript: %s", currentJob.Text)
				}
				
				return // Test passed
			}
		}
	}
}