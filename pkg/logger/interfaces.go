package logger

import "time"

// Interface -.
type Interface interface {
	With(args ...any) Interface
	Debug(message interface{}, args ...interface{})
	Info(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Error(message interface{}, args ...interface{})
	Fatal(message interface{}, args ...interface{})
}

type ChainableLogContext[T any] interface {
	Str(key, val string) T

	Int(key string, val int) T
	Int8(key string, val int8) T
	Int16(key string, val int16) T
	Int32(key string, val int32) T
	Int64(key string, val int64) T
	Uint(key string, val uint) T
	Uint8(key string, val uint8) T
	Uint16(key string, val uint16) T
	Uint32(key string, val uint32) T
	Uint64(key string, val uint64) T
	Float32(key string, val float32) T
	Float64(key string, val float64) T

	Bool(key string, val bool) T

	Time(key string, val time.Time) T
	Dur(key string, val time.Duration) T

	Err(err error) T
}
