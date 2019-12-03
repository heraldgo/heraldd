package util

type loggerI interface {
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
}

// BaseLogger is a basic struct implement LoggerSetter
type BaseLogger struct {
	logger loggerI
}

// Debugf log debug output
func (l *BaseLogger) Debugf(f string, v ...interface{}) {
	if l.logger != nil {
		l.logger.Debugf(f, v...)
	}
}

// Infof log info output
func (l *BaseLogger) Infof(f string, v ...interface{}) {
	if l.logger != nil {
		l.logger.Infof(f, v...)
	}
}

// Warnf log warn output
func (l *BaseLogger) Warnf(f string, v ...interface{}) {
	if l.logger != nil {
		l.logger.Warnf(f, v...)
	}
}

// Errorf log error output
func (l *BaseLogger) Errorf(f string, v ...interface{}) {
	if l.logger != nil {
		l.logger.Errorf(f, v...)
	}
}

// SetLogger will set logger
func (l *BaseLogger) SetLogger(logger interface{}) {
	loggerValue, ok := logger.(loggerI)
	if ok {
		l.logger = loggerValue
	}
}
