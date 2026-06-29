package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// OfflineUploader uploads audio files to the offline server API and returns the transcribed text.
type OfflineUploader struct {
	APIToken string
	APIURL   string
	Timeout  time.Duration
}

// OfflineUploadResult holds the response from the offline server.
type OfflineUploadResult struct {
	Status          string  `json:"status"`
	Text            string  `json:"text"`
	DurationSeconds float64 `json:"duration_seconds"`
	ProcessingTime  float64 `json:"processing_time"`
	ChunksProcessed int     `json:"chunks_processed"`
	Errors          any     `json:"errors"`
	Message         any     `json:"message"`
}

// UploadFile uploads the specified file using multipart/form-data to the offline server.
// Returns the text result or error string.
func (o *OfflineUploader) UploadFile(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	fileWriter, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	_, err = io.Copy(fileWriter, f)
	if err != nil {
		return "", fmt.Errorf("failed to write file data: %w", err)
	}
	writer.Close()

	endpoint := fmt.Sprintf("%s/api/login", o.APIURL)
	req, err := http.NewRequest(http.MethodPost, endpoint, &body)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+o.APIToken)

	client := http.Client{Timeout: o.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result OfflineUploadResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse server response: %w", err)
	}

	if result.Status != "success" {
		errStr := "unsuccessful upload"
		if result.Errors != nil {
			errStr = fmt.Sprintf("%v", result.Errors)
		}
		return "", fmt.Errorf("upload failed: %s", errStr)
	}

	return result.Text, nil
}
