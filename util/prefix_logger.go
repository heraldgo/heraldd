package util

import (
	"github.com/heraldgo/herald"
)

// PrefixLogger automatically add prefix for logger
type PrefixLogger struct {
	Logger herald.Logger
	Prefix string
}

// Debugf log debug output
func (l *PrefixLogger) Debugf(f string, v ...interface{}) {
	if l.Logger != nil {
		l.Logger.Debugf(l.Prefix+" "+f, v...)
	}
}

// Infof log info output
func (l *PrefixLogger) Infof(f string, v ...interface{}) {
	if l.Logger != nil {
		l.Logger.Infof(l.Prefix+" "+f, v...)
	}
}

// Warnf log warn output
func (l *PrefixLogger) Warnf(f string, v ...interface{}) {
	if l.Logger != nil {
		l.Logger.Warnf(l.Prefix+" "+f, v...)
	}
}

// Errorf log error output
func (l *PrefixLogger) Errorf(f string, v ...interface{}) {
	if l.Logger != nil {
		l.Logger.Errorf(l.Prefix+" "+f, v...)
	}
}
