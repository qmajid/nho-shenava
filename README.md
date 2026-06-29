# Audio Recorder

A cross-platform audio recording application that continuously records from the microphone, segments audio into configurable time intervals, and uploads segments to a transcription service via HTTP API with retry and backoff support.

## Features

- **Microphone Recording**: Uses PortAudio for cross-platform audio capture
- **Configurable Segments**: Segment duration configurable via YAML (10-600 seconds)
- **Concurrent Upload**: Multiple worker threads for parallel uploads
- **Retry with Backoff**: Exponential backoff retry on failure
- **Graceful Shutdown**: Clean shutdown on Ctrl+C
- **Cross-Platform**: Works on Linux, macOS, and Raspberry Pi

## Requirements

- Go 1.21 or later
- PortAudio library
- libsndfile (for WAV generation)

### Linux (Debian/Ubuntu)

```bash
sudo apt-get install portaudio libsndfile libsndfile-dev pkg-config
```

### macOS

```bash
brew install portaudio libsndfile
```

### Raspberry Pi

```bash
sudo apt-get install portaudio libsndfile libsndfile-dev
```

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/audio-recorder.git
cd audio-recorder

# Build
make build

# Run
./audio-recorder
```

### Using Docker

```bash
# Build the image
make docker

# Run the container
make docker-run
```

## Configuration

Configuration is stored in `config.yaml`:

```yaml
# Audio recording settings
audio:
  sample_rate: 16000      # Sample rate in Hz
  channels: 1            # Number of channels (1 = mono)
  segment_duration: 60    # Segment duration in seconds (10-600)

# Transcription server settings
server:
  url: http://localhost:8080/transcribe  # Server URL
  timeout: 30           # Request timeout in seconds

# Upload worker settings
workers:
  count: 2              # Number of concurrent workers
  queue_size: 10         # Upload queue size

# Retry settings
retry:
  max_retries: 5        # Maximum retry attempts
  initial_delay: 1      # Initial delay in seconds
  max_delay: 60         # Maximum delay in seconds

# Shutdown settings
shutdown:
  timeout: 30           # Shutdown timeout in seconds

# Logging
log:
  level: info           # Log level: debug, info, warn, error
```

## Usage

### Basic Usage

```bash
# Build and run
make run

# Or run directly
./audio-recorder
```

### With Custom Config

```bash
./audio-recorder -config /path/to/config.yaml
```

### Command Line Options

- `-config` - Path to configuration file (default: `config.yaml`)

## API Server

The application expects a transcription server that accepts:

- **URL**: Configured in `config.yaml`
- **Method**: POST
- **Content-Type**: multipart/form-data
- **Form Fields**:
  - `audio`: WAV audio file
  - `sample_rate`: Sample rate
  - `channels`: Number of channels
  - `duration`: Duration in seconds

### Example Server Response

The server should return JSON:

```json
{
  "success": true,
  "transcription": "Hello world"
}
```

## Building for Different Platforms

```bash
# Linux (x86_64)
make build-linux

# macOS
make build-darwin

# Raspberry Pi (ARM)
make build-arm

# Or use the pi target
make pi
```

## Docker Usage

```bash
# Build the image
docker build -t audio-recorder .

# Run with audio device access
docker run --rm -it \
  --device /dev/snd:/dev/snd \
  audio-recorder

# Run in background
docker run -d \
  --name audio-recorder \
  --device /dev/snd:/dev/snd \
  audio-recorder
```

## Development

### Running Tests

```bash
make test
```

### Code Structure

```
audio-recorder/
├── main.go           # Main entry point
├── audio/
│   ├── recorder.go  # PortAudio recording
│   └── buffer.go   # Audio buffer management
├── network/
│   └── uploader.go # HTTP upload with retry
├── config/
│   └── config.go   # YAML configuration
└── utils/
    ├── logger.go   # Logging
    └── signal.go  # Signal handling
```

## License

MIT License - see LICENSE file for details