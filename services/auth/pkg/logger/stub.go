package logger

// Implementation if stub logger which implements logger interface but does not do anything.
// Used for testing purposes

type StubLogger struct{}

func NewStub() *StubLogger {
	return &StubLogger{}
}

// Debug -.
func (l *StubLogger) Debug(message interface{}, args ...interface{}) {}

// Info -.
func (l *StubLogger) Info(message string, args ...interface{}) {}

// Warn -.
func (l *StubLogger) Warn(message string, args ...interface{}) {}

// Error -.
func (l *StubLogger) Error(message interface{}, args ...interface{}) {}

// Fatal -.
func (l *StubLogger) Fatal(message interface{}, args ...interface{}) {}
