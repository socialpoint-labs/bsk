package logx

import "io"

// Option is the common type of functions that set options
type Option func(*options)

type options struct {
	marshaler               Marshaler
	writer                  io.Writer
	level                   Level
	withoutTime             bool
	withoutFileInfo         bool
	additionalFileSkipLevel int
}

// MarshalerOpt is an option that changes the log marshaler.
func MarshalerOpt(m Marshaler) Option {
	return func(o *options) {
		o.marshaler = m
	}
}

// WriterOpt is an option that changes the log writer.
func WriterOpt(w io.Writer) Option {
	return func(o *options) {
		o.writer = w
	}
}

// LevelOpt is an option that changes the log level.
func LevelOpt(l Level) Option {
	return func(o *options) {
		o.level = l
	}
}

// WithoutTimeOpt is an option  that removes time logging for testing purposes.
func WithoutTimeOpt() Option {
	return func(o *options) {
		o.withoutTime = true
	}
}

// WithoutFileInfo is an option that disables logging the file and line where the log was called.
func WithoutFileInfo() Option {
	return func(o *options) {
		o.withoutFileInfo = true
	}
}

// AdditionalFileSkipLevel is an option that lets you go more levels up to find the file & line doing the log.
func AdditionalFileSkipLevel(l int) Option {
	return func(o *options) {
		o.additionalFileSkipLevel = l
	}
}
