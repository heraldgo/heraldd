package util

import (
	"strings"
)

type loggerI interface {
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
}

// BaseLogger is a basic struct implement LoggerSetter
type BaseLogger struct {
	logger loggerI
	prefix string
}

// Debugf log debug output
func (l *BaseLogger) Debugf(f string, v ...interface{}) {
	if l.logger != nil {
		l.logger.Debugf(l.prefix+f, v...)
	}
}

// Infof log info output
func (l *BaseLogger) Infof(f string, v ...interface{}) {
	if l.logger != nil {
		l.logger.Infof(l.prefix+f, v...)
	}
}

// Warnf log warn output
func (l *BaseLogger) Warnf(f string, v ...interface{}) {
	if l.logger != nil {
		l.logger.Warnf(l.prefix+f, v...)
	}
}

// Errorf log error output
func (l *BaseLogger) Errorf(f string, v ...interface{}) {
	if l.logger != nil {
		l.logger.Errorf(l.prefix+f, v...)
	}
}

// SetLogger will set logger
func (l *BaseLogger) SetLogger(logger interface{}) {
	loggerValue, ok := logger.(loggerI)
	if ok {
		l.logger = loggerValue
	}
}

// SetLoggerPrefix will set logger
func (l *BaseLogger) SetLoggerPrefix(prefix string) {
	l.prefix = strings.TrimSpace(prefix) + " "
}
