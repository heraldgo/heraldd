package main

import (
	"github.com/heraldgo/herald"
)

type prefixLogger struct {
	logger herald.Logger
	prefix string
}

// Debugf log debug output
func (l *prefixLogger) Debugf(f string, v ...interface{}) {
	if l.logger != nil {
		l.logger.Debugf(l.prefix+f, v...)
	}
}

// Infof log info output
func (l *prefixLogger) Infof(f string, v ...interface{}) {
	if l.logger != nil {
		l.logger.Infof(l.prefix+f, v...)
	}
}

// Warnf log warn output
func (l *prefixLogger) Warnf(f string, v ...interface{}) {
	if l.logger != nil {
		l.logger.Warnf(l.prefix+f, v...)
	}
}

// Errorf log error output
func (l *prefixLogger) Errorf(f string, v ...interface{}) {
	if l.logger != nil {
		l.logger.Errorf(l.prefix+f, v...)
	}
}
