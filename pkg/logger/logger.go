package logger

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

// Interface -.
type Interface interface {
	Debug(message interface{}, args ...interface{})
	Info(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Error(message interface{}, args ...interface{})
	Fatal(message interface{}, args ...interface{})
}

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

	skipFrameCount := 3
	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		CallerWithSkipFrameCount(zerolog.CallerSkipFrameCount + skipFrameCount).
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

	if len(args) == 0 || len(args)%2 != 0 {
		logEvent.Msg(messageContent)
		return
	}

	for i := 0; i < len(args)-1; i += 2 {
		_key := args[i]
		_value := args[i+1]

		key, ok := _key.(string)
		if !ok {
			panic(fmt.Errorf("Log argument key should always be a string, got: %T", key))
		}

		switch val := _value.(type) {
		case string:
			logEvent = logEvent.Str(key, val)
		case int:
			logEvent = logEvent.Int(key, val)
		case float64:
			logEvent = logEvent.Float64(key, val)
		case error:
			logEvent = logEvent.Str(key, val.Error())
		default:
			logEvent = logEvent.Str(key, fmt.Sprintf("%+v", val))
		}
	}

	logEvent.Msg(messageContent)
}
