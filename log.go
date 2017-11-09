package main

import (
	"fmt"
	"io"
)

type logger interface {
	WithField(key string, value interface{}) *logger
}

// Log is wrapper for io.Writer with additional Field method
type Log struct {
	Writer io.Writer
	Field  map[string]interface{}
}

// WithField sets prefix
func (l *Log) WithField(key string, value interface{}) *Log {
	field := map[string]interface{}{
		key: value,
	}

	return &Log{Field: field,
		Writer: l.Writer,
	}

}

// Write ads prefix to the end of output message and calls Write method
// from io.Writer package
func (l *Log) Write(p []byte) (n int, err error) {
	var msg = p

	// TODO: do it more graceful
	for key, value := range l.Field {
		field := fmt.Sprintf(" \033[34m%s\033[33m %25s\033[0m", key, value.(string))
		msgWithField := append(p[:len(p)-1], []byte(field+"\n")...)
		msg = msgWithField
		break
	}

	return l.Writer.Write(msg)
}
