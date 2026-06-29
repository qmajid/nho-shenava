package utils

// Level represents log level
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// Logger interface for logging
type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	SetLevel(level Level)
	GetLevel() Level
}

// LogLevelMap maps string to Level
var LogLevelMap = map[string]Level{
	"debug": DebugLevel,
	"info":  InfoLevel,
	"warn":  WarnLevel,
	"error": ErrorLevel,
}

// LevelMap maps Level to string
var LevelMap = map[Level]string{
	DebugLevel: "DEBUG",
	InfoLevel:  "INFO",
	WarnLevel:  "WARN",
	ErrorLevel: "ERROR",
}

func DefaultLogLevel(cfgLogLevel string) Level {
	logLevel := InfoLevel
	if cfgLogLevel == "debug" {
		logLevel = DebugLevel
	} else if cfgLogLevel == "warn" {
		logLevel = WarnLevel
	} else if cfgLogLevel == "error" {
		logLevel = ErrorLevel
	}
	return logLevel
}
