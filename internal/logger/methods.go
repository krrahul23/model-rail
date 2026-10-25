package logger

// With returns a child logger that always includes the supplied fields.
func (l *Logger) With(fields ...Field) (child *Logger) {
	child = &Logger{base: l.baseLogger().With(fields...)}
	return
}

// Named returns a child logger tagged with the given component name.
func (l *Logger) Named(name string) (child *Logger) {
	child = &Logger{base: l.baseLogger().Named(name)}
	return
}

// Sync flushes buffered log entries. Callers may ignore platform-specific
// errors (for example, syncing a terminal on macOS).
func (l *Logger) Sync() (err error) {
	err = l.baseLogger().Sync()
	return
}

// Debug logs at debug level.
func (l *Logger) Debug(msg string, fields ...Field) {
	l.baseLogger().Debug(msg, fields...)
}

// Info logs at info level.
func (l *Logger) Info(msg string, fields ...Field) {
	l.baseLogger().Info(msg, fields...)
}

// Warn logs at warn level.
func (l *Logger) Warn(msg string, fields ...Field) {
	l.baseLogger().Warn(msg, fields...)
}

// Error logs at error level. zap attaches a stack trace at this level.
func (l *Logger) Error(msg string, fields ...Field) {
	l.baseLogger().Error(msg, fields...)
}
