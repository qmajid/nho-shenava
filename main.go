package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"audio-recorder/audio"
	"audio-recorder/config"
	"audio-recorder/network"
	"audio-recorder/utils"
)

func main() {
	// Get log level from config
	cfg := config.NewConfig()

	// Create logger
	logger := utils.NewTerminalLogger(utils.DefaultLogLevel(cfg.Log.Level))
	logger.Info("Audio Recorder starting...")

	// Load configuration
	if err := cfg.LoadConfig("config.yaml"); err != nil {
		logger.Warn(fmt.Sprintf("Failed to load config: %v, using defaults", err))
		os.Exit(1)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		logger.Error(fmt.Sprintf("Invalid configuration: %v", err))
		os.Exit(1)
	}

	if cfg.Audio.OfflineTest {
		testOffline(logger, *cfg)
	}

	// Initialize signal handler with logger
	sigHandler := utils.NewSignalHandler(logger)
	if err := sigHandler.Init(); err != nil {
		logger.Error(fmt.Sprintf("Failed to initialize signal handler: %v", err))
		os.Exit(1)
	}

	// Create audio buffer
	buffer := audio.NewAudioBuffer(
		cfg.Audio.SampleRate,
		cfg.Audio.Channels,
		cfg.Audio.SegmentDuration,
	)

	// Create uploader with logger
	uploader := network.NewUploader(*cfg, logger)

	// Start uploader
	if err := uploader.Start(); err != nil {
		logger.Error(fmt.Sprintf("Failed to start uploader: %v", err))
		os.Exit(1)
	}
	defer uploader.Stop()

	// Create recorder with logger
	recorder, err := audio.NewRecorder(
		cfg.Audio.SampleRate,
		cfg.Audio.Channels,
		logger,
	)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create recorder: %v", err))
		os.Exit(1)
	}

	// Set up audio callback
	recorder.SetCallback(func(samples []int16) {
		if err := buffer.Append(samples); err != nil {
			logger.Error(fmt.Sprintf("Buffer error: %v", err))
			return
		}

		// Send data every 10 seconds if buffer contains data
		staticTickerOnce := false
		var uploadTicker *time.Ticker
		if !staticTickerOnce {
			uploadTicker = time.NewTicker(10 * time.Second)
			staticTickerOnce = true
			go func() {
				for range uploadTicker.C {
					if !buffer.IsEmpty() {
						seg := buffer.GetSegment()
						if seg != nil {
							logger.Info(fmt.Sprintf("Uploading segment: duration=%.2fs", seg.Duration()))
							if transcribe, err := uploader.Upload(seg); err != nil {
								logger.Error(fmt.Sprintf("Upload failed: %v", err))
							} else {
								logger.Info("successful transcribe")
								fmt.Printf("transcribe:\n%v\n", transcribe)
							}

							buffer.Reset()
						}
					}
				}
			}()
		}
	})

	// Start recording
	if err := recorder.Start(); err != nil {
		logger.Error(fmt.Sprintf("Failed to start recording: %v", err))
		os.Exit(1)
	}
	defer recorder.Stop()

	logger.Info("Recording started. Press Ctrl+C to stop.")

	// Wait for shutdown signal
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Monitor for shutdown
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if sigHandler.IsShutdownRequested() {
					logger.Info("Shutdown requested, stopping...")
					cancel()
					return
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// Wait for shutdown
	<-ctx.Done()

	// Graceful shutdown
	logger.Info("Performing graceful shutdown...")

	// Flush remaining buffer
	// if !buffer.IsEmpty() {
	// 	seg := buffer.GetSegment()
	// 	if seg != nil && seg.Duration() > 1.0 {
	// 		logger.Info(fmt.Sprintf("Uploading final segment: duration=%.2fs", seg.Duration()))
	// 		uploader.Upload(seg, func(success bool, resp string) {
	// 			if success {
	// 				logger.Info(fmt.Sprintf("Final upload successful: %s", resp))
	// 			} else {
	// 				logger.Error(fmt.Sprintf("Final upload failed: %s", resp))
	// 			}
	// 		})
	// 	}
	// }

	// Wait for uploader to finish with timeout
	// shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Duration(cfg.Shutdown.Timeout)*time.Second)
	// defer shutdownCancel()

	// done := make(chan struct{})
	// go func() {
	// 	uploader.WaitForComplete()
	// 	close(done)
	// }()

	// select {
	// case <-done:
	// 	logger.Info("All uploads completed")
	// case <-shutdownCtx.Done():
	// 	logger.Warn("Shutdown timeout reached, some uploads may be incomplete")
	// }

	logger.Info("Audio Recorder stopped")
}

func testOffline(logger utils.Logger, cfg config.Config) {
	fmt.Printf("start offline mode for file: %v\n", cfg.Audio.TestFile)
	loginResp, err := network.Login(cfg)
	if err != nil {
		logger.Error(fmt.Sprintf("Uploader failed to login: %v", err))
		os.Exit(1)
	}
	fmt.Printf("login successfully...\n")
	ol := network.OfflineUploader{
		APIToken: loginResp.AccessToken,
		APIURL:   cfg.Server.URL,
		Timeout:  cfg.Server.TranscribeTimeout,
	}
	result, err := ol.UploadFile(cfg.Audio.TestFile)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	fmt.Printf("transcribe result is: \n%v\n", result)
	os.Exit(0)
}
