package run

import (
	"context"
	"time"
)

// WithDeadline runs the given function passing it a context with the deadline adjusted
// to be no later than the provided deadline.  The calling function should respect
// the context.
func WithDeadline(deadline time.Time, fn func(context.Context) error) func(context.Context) error {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithDeadline(ctx, deadline)
		defer cancel()
		return fn(ctx)
	}
}
