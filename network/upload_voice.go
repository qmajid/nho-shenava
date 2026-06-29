package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"audio-recorder/audio"
)

// VoiceUploader uploads captured voice from microphone to the offline server API.
// type VoiceUploader struct {
// 	APIToken string
// 	APIURL   string
// 	Timeout  time.Duration
// }

// UploadResult holds the response from the offline server.
type UploadResult struct {
	Status          string  `json:"status"`
	Text            string  `json:"text"`
	DurationSeconds float64 `json:"duration_seconds"`
	ProcessingTime  float64 `json:"processing_time"`
	ChunksProcessed int     `json:"chunks_processed"`
	Errors          any     `json:"errors"`
	Message         any     `json:"message"`
}

// Upload uploads an AudioSegment to the offline server.
// Returns the transcribed text or error.
func (v *Uploader) Upload(segment *audio.AudioSegment) (string, error) {
	if segment == nil {
		return "", fmt.Errorf("segment is nil")
	}

	// Convert to WAV
	wavData, err := segment.ToWAV()
	if err != nil {
		return "", fmt.Errorf("failed to convert to WAV: %w", err)
	}

	return v.uploadData(wavData, segment.SampleRate(), segment.Channels(), segment.Duration())
}

// uploadData handles the actual multipart upload.
func (v *Uploader) uploadData(wavData []byte, sampleRate, channels int, duration float64) (string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Create form file field
	part, err := writer.CreateFormFile("file", "voice.wav")
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	// Write WAV data
	if _, err := part.Write(wavData); err != nil {
		return "", fmt.Errorf("failed to write audio data: %w", err)
	}

	// Add metadata fields
	if err := writer.WriteField("sample_rate", fmt.Sprintf("%d", sampleRate)); err != nil {
		return "", fmt.Errorf("failed to write sample_rate: %w", err)
	}
	if err := writer.WriteField("channels", fmt.Sprintf("%d", channels)); err != nil {
		return "", fmt.Errorf("failed to write channels: %w", err)
	}
	if err := writer.WriteField("duration", fmt.Sprintf("%.2f", duration)); err != nil {
		return "", fmt.Errorf("failed to write duration: %w", err)
	}

	writer.Close()

	// Create request
	endpoint := fmt.Sprintf("%s/api/transcribe", v.cfg.Server.URL)
	req, err := http.NewRequest(http.MethodPost, endpoint, &body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+"toke")

	// Send request
	client := http.Client{Timeout: v.cfg.Server.TranscribeTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var result UploadResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check status
	if result.Status != "success" {
		errStr := "unknown error"
		if result.Errors != nil {
			errStr = fmt.Sprintf("%v", result.Errors)
		}
		return "", fmt.Errorf("upload failed: %s", errStr)
	}

	return result.Text, nil
}

// NewVoiceUploader creates a new VoiceUploader with the given config.
// func NewVoiceUploader(apiToken, apiURL string, timeout time.Duration) *VoiceUploader {
// 	return &VoiceUploader{
// 		APIToken: apiToken,
// 		APIURL:   apiURL,
// 		Timeout:  timeout,
// 	}
// }
