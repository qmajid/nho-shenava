package audio

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

// AudioSegment represents a segment of audio data
type AudioSegment struct {
	data       []int16
	timestamp  time.Time
	sampleRate int
	channels   int
}

// NewAudioSegment creates a new audio segment
func NewAudioSegment(sampleRate, channels int) *AudioSegment {
	return &AudioSegment{
		data:       make([]int16, 0),
		timestamp:  time.Now(),
		sampleRate: sampleRate,
		channels:   channels,
	}
}

// Duration returns the duration of the segment in seconds
func (s *AudioSegment) Duration() float64 {
	if s.sampleRate == 0 {
		return 0
	}
	return float64(len(s.data)) / float64(s.sampleRate*s.channels)
}

// Data returns the raw audio data
func (s *AudioSegment) Data() []int16 {
	return s.data
}

// SampleRate returns the sample rate
func (s *AudioSegment) SampleRate() int {
	return s.sampleRate
}

// Channels returns the number of channels
func (s *AudioSegment) Channels() int {
	return s.channels
}

// Timestamp returns the segment timestamp
func (s *AudioSegment) Timestamp() time.Time {
	return s.timestamp
}

// ToWAV converts the segment to WAV format
func (s *AudioSegment) ToWAV() ([]byte, error) {
	if len(s.data) == 0 {
		return nil, errors.New("no audio data")
	}

	// WAV header size
	const headerSize = 44

	// Calculate sizes
	sampleBytes := len(s.data) * 2 // 2 bytes per sample
	dataSize := sampleBytes
	fileSize := dataSize + headerSize - 8

	// Create buffer
	buf := new(bytes.Buffer)
	buf.Grow(headerSize + dataSize)

	// RIFF header
	buf.WriteString("RIFF")
	binary.Write(buf, binary.LittleEndian, uint32(fileSize))
	buf.WriteString("WAVE")

	// fmt chunk
	buf.WriteString("fmt ")
	binary.Write(buf, binary.LittleEndian, uint32(16)) // Chunk size
	binary.Write(buf, binary.LittleEndian, uint16(1))  // Audio format (PCM)
	binary.Write(buf, binary.LittleEndian, uint16(s.channels))
	binary.Write(buf, binary.LittleEndian, uint32(s.sampleRate))
	binary.Write(buf, binary.LittleEndian, uint32(s.sampleRate*s.channels*2)) // Byte rate
	binary.Write(buf, binary.LittleEndian, uint16(s.channels*2))              // Block align
	binary.Write(buf, binary.LittleEndian, uint16(16))                        // Bits per sample

	// data chunk
	buf.WriteString("data")
	binary.Write(buf, binary.LittleEndian, uint32(dataSize))

	// Write audio data
	for _, sample := range s.data {
		binary.Write(buf, binary.LittleEndian, sample)
	}

	return buf.Bytes(), nil
}

// AudioBuffer manages audio data buffering
type AudioBuffer struct {
	sampleRate     int
	channels       int
	segmentSeconds time.Duration
	data           []int16
	maxSamples     int
}

// NewAudioBuffer creates a new audio buffer
func NewAudioBuffer(sampleRate, channels int, segmentSeconds time.Duration) *AudioBuffer {
	maxSamples := sampleRate * channels * int(segmentSeconds.Seconds())
	return &AudioBuffer{
		sampleRate:     sampleRate,
		channels:       channels,
		segmentSeconds: segmentSeconds,
		data:           make([]int16, 0, maxSamples),
		maxSamples:     maxSamples,
	}
}

// Append adds samples to the buffer
func (b *AudioBuffer) Append(samples []int16) error {
	// If we would overflow, remove oldest samples
	needed := len(b.data) + len(samples)
	if needed > b.maxSamples {
		// Remove oldest samples to make room
		remove := needed - b.maxSamples
		if remove < len(b.data) {
			b.data = b.data[remove:]
		} else {
			b.data = b.data[:0]
		}
	}

	b.data = append(b.data, samples...)
	return nil
}

// IsFull returns true if the buffer has a complete segment
func (b *AudioBuffer) IsFull() bool {
	return len(b.data) >= b.maxSamples
}

// IsEmpty returns true if the buffer is empty
func (b *AudioBuffer) IsEmpty() bool {
	return len(b.data) == 0
}

// GetSegment returns and clears the current segment
func (b *AudioBuffer) GetSegment() *AudioSegment {
	if len(b.data) == 0 {
		return nil
	}

	segment := NewAudioSegment(b.sampleRate, b.channels)
	segment.data = make([]int16, len(b.data))
	copy(segment.data, b.data)
	segment.timestamp = time.Now()

	return segment
}

// Reset clears the buffer
func (b *AudioBuffer) Reset() {
	b.data = b.data[:0]
}

// Size returns current buffer size in samples
func (b *AudioBuffer) Size() int {
	return len(b.data)
}

// String returns a string representation
func (b *AudioBuffer) String() string {
	return fmt.Sprintf("AudioBuffer{samples=%d/%d, duration=%.2fs}",
		len(b.data), b.maxSamples, b.Duration())
}

// Duration returns current duration in seconds
func (b *AudioBuffer) Duration() float64 {
	if b.sampleRate == 0 {
		return 0
	}
	return float64(len(b.data)) / float64(b.sampleRate*b.channels)
}
