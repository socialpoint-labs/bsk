// Package logx is a logging package inspired by Sirupsen/logrus and
// uber-common/zap that follows these guidelines:
// https://dave.cheney.net/2015/11/05/lets-talk-about-logging
package logx

import (
	"io"
	"io/ioutil"
	"os"
	"time"
)

// Field is a key/value pair associated to a log.
type Field struct {
	Key   string
	Value interface{}
}

// F returns a new log field with the provided key and value
func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Logging levels
const (
	DebugLevel Level = iota + 1
	InfoLevel
)

// Level type
type Level uint8

func (l Level) String() string {
	switch l {
	case 1:
		return "DEBU"
	case 2:
		return "INFO"
	default:
		return "????"
	}
}

// DefaultMinLevel is the minimum debug level for which the logs will appear.
var DefaultMinLevel = DebugLevel

// a log entry has a message, some fields (optional) and a log level
type entry struct {
	message string
	fields  []Field
	level   Level
	time    *time.Time
}

// Logger defines the log methods Debug and Info as defined in
// and also provides a level getter and a method to add fields to a log.
type Logger interface {
	Debug(string, ...Field)
	Info(string, ...Field)
}

// A Log implements Logger and has a marshaler, a writer and a minimum log level.
type Log struct {
	marshaler   Marshaler
	writer      io.Writer
	level       Level
	withoutTime bool
}

// Debug logs a message at level Debug
func (l *Log) Debug(message string, fields ...Field) {
	if DebugLevel >= l.level {
		l.log(DebugLevel, message, fields...)
	}
}

// Info logs a message at level Info
func (l *Log) Info(message string, fields ...Field) {
	if InfoLevel >= l.level {
		l.log(InfoLevel, message, fields...)
	}
}

func (l *Log) log(level Level, message string, fields ...Field) {
	var t *time.Time
	if !l.withoutTime {
		time := time.Now()
		t = &time
	}
	entry := &entry{
		message: message,
		fields:  fields,
		level:   level,
		time:    t,
	}
	data, err := l.marshaler.Marshal(entry)
	if err == nil {
		_, _ = l.writer.Write(data)
	}
	// @TODO log the marshaling has failed?
}

// DefaultWriter is the writer default to all loggers
var DefaultWriter = os.Stdout

// NewLogstash creates a new logstash compatible logger
func NewLogstash(channel, product, application string, opts ...Option) *Log {
	options := &options{}
	for _, opt := range opts {
		opt(options)
	}

	if options.marshaler == nil {
		options.marshaler = NewLogstashMarshaler(channel, product, application)
	}
	if options.writer == nil {
		options.writer = DefaultWriter
	}
	if options.level == 0 {
		options.level = DefaultMinLevel
	}

	return loggerFromOptions(options)
}

// New creates a basic logger with the default values.
func New(opts ...Option) *Log {
	options := &options{}
	for _, opt := range opts {
		opt(options)
	}

	if options.marshaler == nil {
		options.marshaler = new(HumanMarshaler)
	}
	if options.writer == nil {
		options.writer = DefaultWriter
	}
	if options.level == 0 {
		options.level = DefaultMinLevel
	}

	return loggerFromOptions(options)
}

// NewDummy creates a logger for testing purposes.
func NewDummy(opts ...Option) *Log {
	options := &options{}
	for _, opt := range opts {
		opt(options)
	}

	if options.marshaler == nil {
		options.marshaler = new(DummyMarshaler)
	}
	if options.writer == nil {
		options.writer = ioutil.Discard
	}
	if options.level == 0 {
		options.level = DefaultMinLevel
	}

	return loggerFromOptions(options)
}

func loggerFromOptions(opts *options) *Log {
	return &Log{
		opts.marshaler,
		opts.writer,
		opts.level,
		opts.withoutTime,
	}
}
