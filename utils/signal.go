package utils

import (
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
)

// SignalHandler handles OS signals for graceful shutdown
type SignalHandler struct {
	mu                sync.RWMutex
	shutdownRequested atomic.Bool
	signalReceived    int
	sigChan           chan os.Signal
	once              sync.Once
	logger            Logger
}

// NewSignalHandler creates a new signal handler
func NewSignalHandler(logger Logger) *SignalHandler {
	return &SignalHandler{
		logger:  logger,
		sigChan: make(chan os.Signal, 1),
	}
}

// Init initializes the signal handler
func (h *SignalHandler) Init() error {
	signal.Notify(h.sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			sig, ok := <-h.sigChan
			if !ok {
				return
			}
			h.handleSignal(sig)
		}
	}()

	return nil
}

func (h *SignalHandler) handleSignal(sig os.Signal) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Convert signal to int for storage
	var sigNum int
	switch sig {
	case syscall.SIGINT:
		sigNum = int(syscall.SIGINT)
	case syscall.SIGTERM:
		sigNum = int(syscall.SIGTERM)
	default:
		sigNum = 0
	}

	h.signalReceived = sigNum
	h.shutdownRequested.Store(true)

	// Use the injected logger to print the message
	if h.logger != nil {
		h.logger.Warn("Shutdown signal received")
	}
}

// IsShutdownRequested returns true if shutdown was requested
func (h *SignalHandler) IsShutdownRequested() bool {
	return h.shutdownRequested.Load()
}

// GetSignal returns the signal that triggered shutdown
func (h *SignalHandler) GetSignal() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.signalReceived
}

// Reset resets the handler state (for testing)
func (h *SignalHandler) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.shutdownRequested.Store(false)
	h.signalReceived = 0
}
