package logging

import (
	"fmt"
	"log"
	"os"
)

var (
	// InfoLogger logs general information
	InfoLogger *log.Logger
	// ErrorLogger logs error messages
	ErrorLogger *log.Logger
	// DebugLogger logs debug information
	DebugLogger *log.Logger
)

// LogLevel defines the logging verbosity
type LogLevel int

const (
	// LevelInfo logs only important information
	LevelInfo LogLevel = iota
	// LevelDebug logs detailed debugging information
	LevelDebug
)

// InitLogger sets up loggers with different outputs and prefixes
func InitLogger(level LogLevel) {
	// Standard output for info and debug logs
	stdOut := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

	// Error logs go to stderr
	stdErr := log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	InfoLogger = stdOut
	ErrorLogger = stdErr

	// Only create debug logger if debug level is set
	if level == LevelDebug {
		DebugLogger = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	}
}

// Info logs informational messages
func Info(format string, v ...interface{}) {
	if InfoLogger != nil {
		InfoLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

// Error logs error messages
func Error(format string, v ...interface{}) {
	if ErrorLogger != nil {
		ErrorLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

// Debug logs debug messages (only when debug level is set)
func Debug(format string, v ...interface{}) {
	if DebugLogger != nil {
		DebugLogger.Output(2, fmt.Sprintf(format, v...))
	}
}
