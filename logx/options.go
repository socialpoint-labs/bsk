package logx

import "io"

// Option is the common type of functions that set options
type Option func(*options)

type options struct {
	marshaler     Marshaler
	writer        io.Writer
	level         Level
	withoutTime   bool
	fileSkipLevel int
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

// FileSkipLevel is an option that specify is the number of stack frames to ascend to get the calling file
func FileSkipLevel(l int) Option {
	return func(o *options) {
		o.fileSkipLevel = l
	}
}
