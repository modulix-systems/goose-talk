package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"
)

type LogLevel int8

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// Logger -.
type Logger struct {
	logger *zerolog.Logger
}

var _ Interface = (*Logger)(nil)

// New -.
func New(level LogLevel) *Logger {
	zerolog.SetGlobalLevel(zerolog.Level(level))

	var outputWriter io.Writer = os.Stdout
	if level == DebugLevel {
		outputWriter = zerolog.ConsoleWriter{Out: os.Stdout}
	}

	logger := zerolog.New(outputWriter).
		With().
		Timestamp().
		Logger()

	return &Logger{
		logger: &logger,
	}
}

// Debug -.
func (l *Logger) Debug(message interface{}, args ...interface{}) {
	l.log(l.logger.Debug(), message, args...)
}

// Info -.
func (l *Logger) Info(message string, args ...interface{}) {
	l.log(l.logger.Info(), message, args...)
}

// Warn -.
func (l *Logger) Warn(message string, args ...interface{}) {
	l.log(l.logger.Warn(), message, args...)
}

// Error -.
func (l *Logger) Error(message interface{}, args ...interface{}) {
	l.log(l.logger.Error(), message, args...)
}

// Fatal -.
func (l *Logger) Fatal(message interface{}, args ...interface{}) {
	l.log(l.logger.Fatal(), message, args...)
}

func (l *Logger) log(logEvent *zerolog.Event, message interface{}, args ...interface{}) {
	var messageContent string

	switch msg := message.(type) {
	case error:
		messageContent = msg.Error()
	case string:
		messageContent = msg
	default:
		panic(fmt.Sprintf("message %v has unknown type %v", message, msg))
	}

	logEvent = attachContextFields(logEvent, args...)
	logEvent.Msg(messageContent)
}

func (l *Logger) With(args ...any) Interface {
	logger := attachContextFields(l.logger.With(), args...).Logger()
	return &Logger{&logger}
}
