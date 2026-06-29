package audio

import (
	"fmt"
	"sync"
	"time"

	"audio-recorder/utils"

	"github.com/gordonklaus/portaudio"
)

// AudioCallback is called with recorded audio samples
type AudioCallback func(samples []int16)

// Recorder handles audio recording from microphone
type Recorder struct {
	sampleRate      int
	channels        int
	stream          *portaudio.Stream
	callback        AudioCallback
	stopped         bool
	mu              sync.Mutex
	logger          utils.Logger
	buffer          []int16
	framesPerBuffer int
}

// NewRecorder creates a new audio recorder
func NewRecorder(sampleRate, channels int, logger utils.Logger) (*Recorder, error) {
	if err := portaudio.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize portaudio: %v", err)
	}

	framesPerBuffer := 1024
	return &Recorder{
		sampleRate:      sampleRate,
		channels:        channels,
		stopped:         true,
		logger:          logger,
		buffer:          make([]int16, framesPerBuffer*channels),
		framesPerBuffer: framesPerBuffer,
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

	// Allocate buffer
	// r.buffer = make([]int16, r.framesPerBuffer*r.channels)

	// Create PortAudio stream
	stream, err := portaudio.OpenDefaultStream(
		r.channels,
		0,
		float64(r.sampleRate),
		r.framesPerBuffer,
		&r.buffer,
	)
	if err != nil {
		return fmt.Errorf("failed to open stream: %v", err)
	}
	r.stream = stream

	if err := r.stream.Start(); err != nil {
		r.stream.Close()
		return fmt.Errorf("failed to start stream: %v", err)
	}

	r.stopped = false

	go r.readLoop()

	return nil
}

func (r *Recorder) readLoop() {
	for {
		r.mu.Lock()
		stopped := r.stopped
		callback := r.callback
		stream := r.stream
		bufferCopy := make([]int16, len(r.buffer))
		copy(bufferCopy, r.buffer)
		r.mu.Unlock()

		if stopped {
			break
		}

		if callback != nil && stream != nil {
			// Read samples into buffer
			err := stream.Read()
			if err == nil {
				// Copy buffer to not overwrite
				data := make([]int16, len(r.buffer))
				copy(data, r.buffer)

				// If we have fewer than framesPerBuffer, slice accordingly
				callback(data)
			} else if err != portaudio.InputOverflowed {
				r.logger.Error(fmt.Sprintf("Audio read error: %v", err))
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// Stop stops recording
func (r *Recorder) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stopped {
		return nil
	}

	r.stopped = true

	if r.stream != nil {
		r.stream.Stop()
		r.stream.Close()
		r.stream = nil
	}

	return nil
}

// Close releases recorder resources
func (r *Recorder) Close() error {
	r.Stop()
	portaudio.Terminate()
	return nil
}
