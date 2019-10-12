package util

import (
	"github.com/xianghuzhao/herald"
)

// BaseLogger is a basic struct implement LoggerSetter
type BaseLogger struct {
	Logger herald.Logger
}

// SetLogger will set logger
func (l *BaseLogger) SetLogger(logger herald.Logger) {
	l.Logger = logger
}
