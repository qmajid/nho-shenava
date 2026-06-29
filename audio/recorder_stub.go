//go:build !portaudio
// +build !portaudio

package audio

import (
	"sync"

	"audio-recorder/utils"
)

// AudioCallback is called with recorded audio samples
type AudioCallback func(samples []int16)

// Recorder handles audio recording from microphone
type Recorder struct {
	sampleRate int
	channels  int
	callback AudioCallback
	stopped   bool
	mu        sync.Mutex
	logger    utils.Logger
}

// NewRecorder creates a new audio recorder
func NewRecorder(sampleRate, channels int, logger utils.Logger) (*Recorder, error) {
	return &Recorder{
		sampleRate: sampleRate,
		channels:  channels,
		stopped:   true,
		logger:   logger,
	}, nil
}

// SetCallback sets the callback for audio data
func (r *Recorder) SetCallback(callback AudioCallback) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.callback = callback
}

// Start starts recording
func (r *Recorder) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.stopped {
		return nil
	}

	r.stopped = false
	return nil
}

// Stop stops recording
func (r *Recorder) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stopped {
		return nil
	}

	r.stopped = true
	return nil
}

// Close releases recorder resources
func (r *Recorder) Close() error {
	return r.Stop()
}

// ErrPortAudioNotAvailable is returned when PortAudio is not available
var ErrPortAudioNotAvailable = &RecorderError{Message: "PortAudio not available, build with -tags=portaudio"}

// RecorderError represents a recorder error
type RecorderError struct {
	Message string
}

func (e *RecorderError) Error() string {
	return e.Message
}