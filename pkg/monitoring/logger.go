package monitoring

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

type Logger struct {
	level  LogLevel
	fields map[string]interface{}
}

func NewLogger(level LogLevel) *Logger {
	return &Logger{
		level:  level,
		fields: make(map[string]interface{}),
	}
}

func (l *Logger) WithField(key string, value interface{}) *Logger {
	newLogger := &Logger{
		level:  l.level,
		fields: make(map[string]interface{}),
	}

	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new field
	newLogger.fields[key] = value
	return newLogger
}

func (l *Logger) log(level LogLevel, levelStr string, message string, args ...interface{}) {
	if level < l.level {
		return
	}

	logEntry := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"level":     levelStr,
		"message":   fmt.Sprintf(message, args...),
	}

	for k, v := range l.fields {
		logEntry[k] = v
	}

	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		log.Printf("ERROR: Failed to marshal log entry: %v", err)
		return
	}

	fmt.Fprintln(os.Stdout, string(jsonData))
}

func (l *Logger) Debug(message string, args ...interface{}) {
	l.log(DEBUG, "DEBUG", message, args...)
}

func (l *Logger) Info(message string, args ...interface{}) {
	l.log(INFO, "INFO", message, args...)
}

func (l *Logger) Warn(message string, args ...interface{}) {
	l.log(WARN, "WARN", message, args...)
}

func (l *Logger) Error(message string, args ...interface{}) {
	l.log(ERROR, "ERROR", message, args...)
}
