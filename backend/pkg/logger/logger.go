package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"time"
)

// Log is the global logger instance
var Log *slog.Logger

// Init initializes the global logger with the specified level
func Init(level slog.Level) {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: true, // Source info will be added from the PC we provide
	})
	Log = slog.New(handler)
}

// Error logs an error message with additional attributes
// The source location will point to the caller, not this function
func Error(msg string, err error, attrs ...any) {
	if !Log.Enabled(context.Background(), slog.LevelError) {
		return
	}
	
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Error]
	r := slog.NewRecord(time.Now(), slog.LevelError, msg, pcs[0])
	
	if err != nil {
		r.Add("error", err.Error())
	}
	r.Add(attrs...)
	
	_ = Log.Handler().Handle(context.Background(), r)
}

// Info logs an informational message with additional attributes
func Info(msg string, attrs ...any) {
	if !Log.Enabled(context.Background(), slog.LevelInfo) {
		return
	}
	
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Info]
	r := slog.NewRecord(time.Now(), slog.LevelInfo, msg, pcs[0])
	r.Add(attrs...)
	
	_ = Log.Handler().Handle(context.Background(), r)
}

// Warn logs a warning message with additional attributes
func Warn(msg string, attrs ...any) {
	if !Log.Enabled(context.Background(), slog.LevelWarn) {
		return
	}
	
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Warn]
	r := slog.NewRecord(time.Now(), slog.LevelWarn, msg, pcs[0])
	r.Add(attrs...)
	
	_ = Log.Handler().Handle(context.Background(), r)
}

// Debug logs a debug message with additional attributes
func Debug(msg string, attrs ...any) {
	if !Log.Enabled(context.Background(), slog.LevelDebug) {
		return
	}
	
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Debug]
	r := slog.NewRecord(time.Now(), slog.LevelDebug, msg, pcs[0])
	r.Add(attrs...)
	
	_ = Log.Handler().Handle(context.Background(), r)
}

// With returns a new logger with the given attributes
func With(attrs ...any) *slog.Logger {
	return Log.With(attrs...)
}
