# Audio Recorder & Transcription Service

## Project Overview

- **Project Name**: audio-recorder
- **Type**: CLI Audio Recording Application
- **Core Functionality**: Continuous microphone recording with minute-level segmentation and API upload for transcription
- **Target Users**: Developers and systems requiring automated audio transcription

## Functionality Specification

### Core Features

#### 1. Microphone Recording (PortAudio)
- Use PortAudio library for cross-platform audio capture
- Support default audio input device
- Sample rate: 16000 Hz (optimal for speech recognition)
- Channels: Mono (1 channel)
- Format: 16-bit signed integer (PCM)
- Buffer size: 1024 frames

#### 2. Segment Duration Configuration
- Configurable segment length via YAML (default: 60 seconds)
- Support values from 10 seconds to 600 seconds
- Audio buffered in memory, split at segment boundaries

#### 3. Transcription Server Configuration
- Configurable server URL via YAML
- Default: `http://localhost:8080/transcribe`
- HTTP POST with multipart/form-data
- Audio format: WAV (16-bit PCM, mono, 16kHz)
- JSON response handling

#### 4. YAML Configuration
- Configuration file: `config.yaml`
- All runtime parameters configurable
- Default values embedded in code

#### 5. Concurrent Upload Workers
- Configurable worker count (default: 2)
- Worker pool using goroutines
- Channel-based queue for pending uploads
- Thread-safe audio segment storage

#### 6. Retry with Backoff
- Exponential backoff retry strategy
- Initial delay: 1 second
- Maximum delay: 60 seconds
- Maximum retries: 5 (configurable)
- Retry on network errors and 5xx responses

#### 7. Graceful Shutdown (Ctrl+C)
- Signal handler for SIGINT/SIGTERM
- Flush pending buffers before exit
- Wait for ongoing uploads to complete
- Timeout: 30 seconds for graceful shutdown

#### 8. Dockerfile
- Multi-stage build for size optimization
- Alpine base with PortAudio
- Cross-compile support for ARM (Raspberry Pi)
- Non-root user for security

#### 9. Makefile
- Build targets: all, clean, install, run
- Platform detection (Linux/macOS)
- Optional cross-compile for ARM
- Test target

#### 10. Cross-Platform Support
- Linux (including Raspberry Pi)
- macOS
- x86_64 and ARM architectures
- Conditional compilation for platform-specific code

#### 11. Ready to Go Build
- Single command build: `make`
- Default config file included
- Binary output: `audio-recorder`

### User Interactions and Flows

1. Start application: `./audio-recorder` or `make run`
2. Application loads config from `config.yaml`
3. Begins continuous recording
4. Every N seconds (configurable), segment is queued for upload
5. Worker pool uploads segments in parallel
6. Ctrl+C triggers graceful shutdown
7. Pending segments are uploaded, then application exits

### Configuration File Format (config.yaml)

```yaml
# Audio recording settings
audio:
  sample_rate: 16000
  channels: 1
  segment_duration: 60  # seconds (10-600)

# Transcription server settings
server:
  url: http://localhost:8080/transcribe
  timeout: 30  # seconds

# Upload worker settings
workers:
  count: 2
  queue_size: 10

# Retry settings
retry:
  max_retries: 5
  initial_delay: 1  # seconds
  max_delay: 60  # seconds

# Shutdown settings
shutdown:
  timeout: 30  # seconds

# Logging
log:
  level: info  # debug, info, warn, error
```

### Edge Cases

- No microphone available: Exit with error message
- Network unavailable: Retry with backoff, continue recording
- Server returns error: Retry, log error after max retries
- Buffer overflow: Drop oldest segments (ring buffer)
- Config file missing: Use defaults, log warning
- Invalid config: Exit with error message

## Acceptance Criteria

1. Application compiles on Linux and macOS
2. Application compiles for ARM (Raspberry Pi)
3. Recording produces valid WAV files
4. Segments are exactly the configured duration
5. Uploads succeed with retry on failure
6. Ctrl+C triggers graceful shutdown within timeout
7. Configuration loads from YAML file
8. All settings are configurable without code changes
9. Dockerfile produces working container
10. Makefile provides all standard targets