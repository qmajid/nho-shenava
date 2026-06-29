package network

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"sync"
	"time"

	"audio-recorder/audio"
	"audio-recorder/config"
	"audio-recorder/utils"
)

// UploadCallback is called when upload completes
type UploadCallback func(success bool, response string)

// Uploader handles uploading audio segments to transcription service
type Uploader struct {
	cfg    config.Config
	logger utils.Logger

	jobs      chan *uploadJob
	workers   []*worker
	waitGroup sync.WaitGroup
	running   bool
	mu        sync.RWMutex
}

type uploadJob struct {
	segment  *audio.AudioSegment
	callback UploadCallback
}

type worker struct {
	id       int
	uploader *Uploader
}

// NewUploader creates a new uploader
func NewUploader(cfg config.Config, logger utils.Logger) *Uploader {
	return &Uploader{
		cfg:    cfg,
		jobs:   make(chan *uploadJob, cfg.Workers.Count*10),
		logger: logger,
	}
}

// Start starts the uploader workers
func (u *Uploader) Start() error {
	if u.running {
		return nil
	}

	// Authenticate with the server using the Login function.
	loginResp, err := Login(u.cfg)
	if err != nil {
		u.logger.Error(fmt.Sprintf("Uploader failed to login: %v", err))
		return err
	}
	u.logger.Info(fmt.Sprintf("Uploader logged in successfully. Token type: %+v", loginResp))
	// Store the access token somewhere if needed (not shown here)

	u.running = true
	u.workers = make([]*worker, u.cfg.Workers.Count)

	for i := 0; i < u.cfg.Workers.Count; i++ {
		w := &worker{
			id:       i,
			uploader: u,
		}
		u.workers[i] = w

		u.waitGroup.Add(1)
		go func() {
			defer u.waitGroup.Done()
			w.run()
		}()
	}

	u.logger.Info(fmt.Sprintf("Started %d upload workers", u.cfg.Workers.Count))
	return nil
}

// Stop stops the uploader
func (u *Uploader) Stop() {
	if !u.running {
		return
	}

	u.running = false
	close(u.jobs)
	u.waitGroup.Wait()
}

// Upload queues an audio segment for upload
// func (u *Uploader) Upload(segment *audio.AudioSegment, callback UploadCallback) {
// 	job := &uploadJob{
// 		segment:  segment,
// 		callback: callback,
// 	}

// 	select {
// 	case u.jobs <- job:
// 		// Job queued
// 	default:
// 		// Queue full, drop the job
// 		u.logger.Warn("Upload queue full, dropping segment")
// 		if callback != nil {
// 			callback(false, "queue full")
// 		}
// 	}
// }

// WaitForComplete waits for all pending uploads to complete
func (u *Uploader) WaitForComplete() {
	// Give a moment for last uploads to be queued
	time.Sleep(100 * time.Millisecond)

	// Wait for all workers to finish
	u.waitGroup.Wait()
}

func (w *worker) run() {
	for {
		select {
		case job, ok := <-w.uploader.jobs:
			if !ok {
				return
			}
			w.processJob(job)
		}
	}
}

func (w *worker) processJob(job *uploadJob) {
	var response string
	success := false

	for attempt := 0; attempt <= w.uploader.cfg.Retry.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate backoff delay
			delay := w.calculateBackoff(attempt)
			w.uploader.logger.Info(fmt.Sprintf("Retry %d/%d after %.1fs", attempt, w.uploader.cfg.Retry.MaxRetries, delay.Seconds()))
			time.Sleep(delay)
		}

		success, response = w.upload(job.segment)
		if success {
			break
		}
	}

	if job.callback != nil {
		job.callback(success, response)
	}
}

func (w *worker) upload(segment *audio.AudioSegment) (bool, string) {
	// Convert to WAV
	wavData, err := segment.ToWAV()
	if err != nil {
		return false, fmt.Sprintf("failed to convert to WAV: %v", err)
	}

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add audio field
	part, err := writer.CreateFormFile("audio", "segment.wav")
	if err != nil {
		return false, fmt.Sprintf("failed to create form file: %v", err)
	}

	if _, err := part.Write(wavData); err != nil {
		return false, fmt.Sprintf("failed to write audio data: %v", err)
	}

	// Add metadata
	if err := writer.WriteField("sample_rate", fmt.Sprintf("%d", segment.SampleRate())); err != nil {
		return false, fmt.Sprintf("failed to write sample_rate: %v", err)
	}
	if err := writer.WriteField("channels", fmt.Sprintf("%d", segment.Channels())); err != nil {
		return false, fmt.Sprintf("failed to write channels: %v", err)
	}
	if err := writer.WriteField("duration", fmt.Sprintf("%.2f", segment.Duration())); err != nil {
		return false, fmt.Sprintf("failed to write duration: %v", err)
	}

	writer.Close()

	// Create request
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(w.uploader.cfg.Server.Timeout)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", w.uploader.cfg.Server.URL, body)
	if err != nil {
		return false, fmt.Sprintf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	client := http.Client{
		Timeout: w.uploader.cfg.Server.TranscribeTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Sprintf("failed to read response: %v", err)
	}

	if resp.StatusCode >= 500 {
		// Server error, retry
		return false, fmt.Sprintf("server error: %d - %s", resp.StatusCode, string(respBody))
	}

	if resp.StatusCode >= 400 {
		// Client error, don't retry
		return false, fmt.Sprintf("client error: %d - %s", resp.StatusCode, string(respBody))
	}

	return true, string(respBody)
}

func (w *worker) calculateBackoff(attempt int) time.Duration {
	delay := float64(w.uploader.cfg.Retry.InitialDelay) * math.Pow(2, float64(attempt-1))
	if delay > float64(w.uploader.cfg.Retry.MaxDelay) {
		delay = float64(w.uploader.cfg.Retry.MaxDelay)
	}
	return time.Duration(delay * float64(time.Second))
}
