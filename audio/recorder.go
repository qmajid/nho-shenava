//go:build portaudio
// +build portaudio

package audio

/*
#cgo darwin pkg-config: portaudio-2.0
#cgo linux,!arm pkg-config: portaudio-2.0
#cgo linux,arm pkg-config: portaudio-2.0
#cgo LDFLAGS: -lportaudio

#include <portaudio.h>
#include <stdlib.h>
#include <string.h>

// streamData holds the audio buffer
typedef struct {
    float *buffer;
    int bufferSize;
    int writePos;
    int channels;
    volatile int recording;
} streamData;

// Global stream data pointer
static streamData *g_streamData = NULL;

// Callback function for PortAudio
static int paCallback(const void *inputBuffer, void *outputBuffer,
                   unsigned long framesPerBuffer,
                   const PaStreamCallbackTimeInfo *timeInfo,
                   PaStreamCallbackFlags statusFlags,
                   void *userData) {
    (void)outputBuffer;
    (void)timeInfo;
    (void)statusFlags;
    (void)userData;

    if (inputBuffer == NULL) {
        return paContinue;
    }

    const float *in = (const float *)inputBuffer;
    streamData *data = g_streamData;

    if (data == NULL || !data->recording) {
        return paContinue;
    }

    unsigned long samplesToCopy = framesPerBuffer * data->channels;
    int samplesAvailable = data->bufferSize - data->writePos;

    if (samplesToCopy > (unsigned long)samplesAvailable) {
        int samplesToKeep = data->writePos - (int)samplesToCopy;
        if (samplesToKeep > 0) {
            memmove(data->buffer, data->buffer + data->writePos - samplesToKeep,
                   samplesToKeep * sizeof(float));
        }
        data->writePos = samplesToKeep < 0 ? 0 : samplesToKeep;
    }

    for (unsigned long i = 0; i < samplesToCopy && data->writePos < data->bufferSize; i++) {
        data->buffer[data->writePos + i] = in[i];
    }
    data->writePos += samplesToCopy;

    if (data->writePos > data->bufferSize) {
        data->writePos = data->bufferSize;
    }

    return paContinue;
}
*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"

	"audio-recorder/utils"
)

// AudioCallback is called with recorded audio samples
type AudioCallback func(samples []int16)

// Recorder handles audio recording from microphone
type Recorder struct {
	sampleRate int
	channels   int
	stream     *C.PaStream
	callback   AudioCallback
	streamData *C.streamData
	stopped    bool
	mu         sync.Mutex
	logger     utils.Logger
}

// NewRecorder creates a new audio recorder
func NewRecorder(sampleRate, channels int, logger utils.Logger) (*Recorder, error) {
	if err := C.Pa_Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize portaudio: %v", C.Pa_GetErrorText(err))
	}

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

	r.streamData = (*C.streamData)(C.malloc(C.sizeof_streamData))
	r.streamData.buffer = (*C.float)(C.malloc(C.size_t(r.sampleRate * r.channels * 2 * 4)))
	r.streamData.bufferSize = C.int(r.sampleRate * r.channels * 2)
	r.streamData.writePos = 0
	r.streamData.channels = C.int(r.channels)
	r.streamData.recording = 1

	C.g_streamData = r.streamData

	inputDevice := C.Pa_GetDefaultInputDevice()
	if inputDevice == C.paNoDevice {
		return fmt.Errorf("no input device available")
	}

	deviceInfo := C.Pa_GetDeviceInfo(inputDevice)
	if deviceInfo == nil {
		return fmt.Errorf("failed to get device info")
	}

	inputParams := C.Pa_StreamParameters{
		device:               inputDevice,
		channelCount:        C.int(r.channels),
		sampleFormat:        C.paFloat32,
		latency:             deviceInfo.defaultLowInputLatency,
		hostApiSpecificStreamInfo: nil,
	}

	var stream *C.PaStream
	err := C.Pa_OpenStream(
		&stream,
		&inputParams,
		nil,
		C.double(r.sampleRate),
		C.ulong(1024),
		C.paClipOff,
		C.PaStreamCallback(C.paCallback),
		nil,
	)
	if err != 0 {
		return fmt.Errorf("failed to open stream: %v", C.Pa_GetErrorText(err))
	}

	r.stream = stream

	if err := C.Pa_StartStream(stream); err != 0 {
		C.Pa_CloseStream(stream)
		return fmt.Errorf("failed to start stream: %v", C.Pa_GetErrorText(err))
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
		streamData := r.streamData
		r.mu.Unlock()

		if stopped {
			break
		}

		if streamData != nil && callback != nil {
			writePos := int(streamData.writePos)
			if writePos > 0 {
				samples := make([]int16, writePos)
				for i := 0; i < writePos; i++ {
					v := float32(streamData.buffer[i])
					if v > 1 {
						v = 1
					}
					if v < -1 {
						v = -1
					}
					samples[i] = int16(v * 32767)
				}

				streamData.writePos = 0

				callback(samples)
			}
		}

		C.Pa_Sleep(10)
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

	if r.streamData != nil {
		r.streamData.recording = 0
	}

	if r.stream != nil {
		C.Pa_StopStream(r.stream)
		C.Pa_CloseStream(r.stream)
		r.stream = nil
	}

	if r.streamData != nil {
		if r.streamData.buffer != nil {
			C.free(unsafe.Pointer(r.streamData.buffer))
		}
		C.free(unsafe.Pointer(r.streamData))
		r.streamData = nil
	}

	return nil
}

// Close releases recorder resources
func (r *Recorder) Close() error {
	r.Stop()
	C.Pa_Terminate()
	return nil
}