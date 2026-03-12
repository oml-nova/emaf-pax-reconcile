package superlog

import (
	"fmt"
	"sync"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

type LogEvent struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       LogLevel               `json:"level"`
	Message     string                 `json:"message"`
	Correlation string                 `json:"correlation,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

type ILogger interface {
	Debug(message string, data map[string]interface{}) *LogEvent
	Info(message string, data map[string]interface{}) *LogEvent
	Warn(message string, data map[string]interface{}) *LogEvent
	Error(message string, data map[string]interface{}) *LogEvent
}

type LogShipper interface {
	Ship(event LogEvent) error
}

type Logger struct {
	shippers []LogShipper
}

var (
	instance *Logger
	once     sync.Once
)

// Log returns the singleton. Call after NewLogger().
func Log() *Logger { return instance }

// NewLogger creates the singleton. Subsequent calls are no-ops (sync.Once).
func NewLogger(shippers ...LogShipper) *Logger {
	once.Do(func() {
		instance = &Logger{shippers: shippers}
	})
	return instance
}

func (l *Logger) log(level LogLevel, message string, data map[string]interface{}) *LogEvent {
	event := LogEvent{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Data:      data,
	}
	for _, shipper := range l.shippers {
		go func() {
			if err := shipper.Ship(event); err != nil {
				fmt.Printf("Error shipping log event: %v\n", err)
			}
		}()
	}
	return &event
}

func (l *Logger) Debug(message string, data map[string]interface{}) *LogEvent {
	return l.log(DEBUG, message, data)
}
func (l *Logger) Info(message string, data map[string]interface{}) *LogEvent {
	return l.log(INFO, message, data)
}
func (l *Logger) Warn(message string, data map[string]interface{}) *LogEvent {
	return l.log(WARN, message, data)
}
func (l *Logger) Error(message string, data map[string]interface{}) *LogEvent {
	return l.log(ERROR, message, data)
}

// Correlate chains a correlation ID onto a LogEvent (fluent API).
func (e *LogEvent) Correlate(correlationID string) *LogEvent {
	e.Correlation = correlationID
	return e
}
