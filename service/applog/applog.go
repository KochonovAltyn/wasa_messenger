// Package applog is a small structured-logging facade built on the standard
// library's log/slog. It replaces the previously used github.com/sirupsen/logrus
// dependency while preserving the subset of the logrus API used throughout this
// project (New, SetOutput, SetLevel, WithError, WithFields, and the leveled
// printf/println-style methods).
package applog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
)

// Level mirrors the logrus level concept for the values this project uses.
type Level int

const (
	// DebugLevel logs everything, including debug messages.
	DebugLevel Level = iota
	// InfoLevel logs informational messages and above.
	InfoLevel
)

// Fields is a map of structured key/value pairs attached to a log entry.
// It is the drop-in equivalent of logrus.Fields.
type Fields map[string]interface{}

// FieldLogger is the interface consumed by the rest of the codebase. It is the
// equivalent of logrus.FieldLogger for the methods actually used here.
type FieldLogger interface {
	WithError(err error) FieldLogger
	WithFields(fields Fields) FieldLogger

	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Warning(args ...interface{})
	Error(args ...interface{})

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})

	Println(args ...interface{})
}

// Logger is the concrete logger. It wraps a *slog.Logger plus a configurable
// minimum level and output, and carries any attached structured fields.
type Logger struct {
	out   io.Writer
	level Level
	sl    *slog.Logger
	attrs []slog.Attr
}

// New creates a Logger writing to stdout at InfoLevel by default.
func New() *Logger {
	l := &Logger{out: os.Stdout, level: InfoLevel}
	l.rebuild()
	return l
}

// SetOutput sets the destination writer for log entries.
func (l *Logger) SetOutput(w io.Writer) {
	l.out = w
	l.rebuild()
}

// SetLevel sets the minimum level that will be emitted.
func (l *Logger) SetLevel(level Level) {
	l.level = level
	l.rebuild()
}

func (l *Logger) slogLevel() slog.Level {
	if l.level == DebugLevel {
		return slog.LevelDebug
	}
	return slog.LevelInfo
}

func (l *Logger) rebuild() {
	handler := slog.NewTextHandler(l.out, &slog.HandlerOptions{Level: l.slogLevel()})
	l.sl = slog.New(handler)
}

// child clones the logger, preserving output/level/handler and adding attrs.
func (l *Logger) child(extra []slog.Attr) *Logger {
	merged := make([]slog.Attr, 0, len(l.attrs)+len(extra))
	merged = append(merged, l.attrs...)
	merged = append(merged, extra...)
	return &Logger{out: l.out, level: l.level, sl: l.sl, attrs: merged}
}

// WithError returns a logger that attaches the given error under the "error" key.
func (l *Logger) WithError(err error) FieldLogger {
	return l.child([]slog.Attr{slog.Any("error", err)})
}

// WithFields returns a logger that attaches the given structured fields.
func (l *Logger) WithFields(fields Fields) FieldLogger {
	extra := make([]slog.Attr, 0, len(fields))
	for k, v := range fields {
		extra = append(extra, slog.Any(k, v))
	}
	return l.child(extra)
}

func (l *Logger) log(level slog.Level, msg string) {
	// slog discards entries below the handler level, so this is safe to call always.
	l.sl.LogAttrs(context.Background(), level, msg, l.attrs...)
}

// Leveled, args-joined methods (logrus print-style).

func (l *Logger) Debug(args ...interface{})   { l.log(slog.LevelDebug, fmt.Sprint(args...)) }
func (l *Logger) Info(args ...interface{})    { l.log(slog.LevelInfo, fmt.Sprint(args...)) }
func (l *Logger) Warn(args ...interface{})    { l.log(slog.LevelWarn, fmt.Sprint(args...)) }
func (l *Logger) Warning(args ...interface{}) { l.log(slog.LevelWarn, fmt.Sprint(args...)) }
func (l *Logger) Error(args ...interface{})   { l.log(slog.LevelError, fmt.Sprint(args...)) }
func (l *Logger) Println(args ...interface{}) { l.log(slog.LevelInfo, fmt.Sprint(args...)) }

// Leveled, formatted methods (logrus printf-style).

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(slog.LevelDebug, fmt.Sprintf(format, args...))
}
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(slog.LevelInfo, fmt.Sprintf(format, args...))
}
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(slog.LevelWarn, fmt.Sprintf(format, args...))
}
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(slog.LevelError, fmt.Sprintf(format, args...))
}
