package sqsx

import "time"

type Option func(*options)

type options struct {
	visibilityTimeout *time.Duration
	waitTime          *time.Duration
}

func WithVisibilityTimeout(d time.Duration) Option {
	return func(opts *options) {
		opts.visibilityTimeout = &d
	}
}

func WithWaitTime(d time.Duration) Option {
	return func(opts *options) {
		opts.waitTime = &d
	}
}
