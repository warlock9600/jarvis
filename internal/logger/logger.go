package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type Level int

const (
	LevelError Level = iota
	LevelWarn
	LevelInfo
	LevelDebug
)

type Logger struct {
	mu    sync.Mutex
	out   io.Writer
	level Level
}

func New(level Level) *Logger {
	return &Logger{out: os.Stderr, level: level}
}

func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *Logger) logf(level Level, label string, format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if level > l.level {
		return
	}
	ts := time.Now().Format(time.RFC3339)
	fmt.Fprintf(l.out, "%s %-5s %s\n", ts, label, fmt.Sprintf(format, args...))
}

func (l *Logger) Errorf(format string, args ...any) {
	l.logf(LevelError, "ERROR", format, args...)
}

func (l *Logger) Warnf(format string, args ...any) {
	l.logf(LevelWarn, "WARN", format, args...)
}

func (l *Logger) Infof(format string, args ...any) {
	l.logf(LevelInfo, "INFO", format, args...)
}

func (l *Logger) Debugf(format string, args ...any) {
	l.logf(LevelDebug, "DEBUG", format, args...)
}
