package run

import (
	"context"
	"time"
)

// Sleep waits for the time duration, or the context is canceled, whichever happens first.
func Sleep(ctx context.Context, d time.Duration) {
	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	<-ctx.Done()
}
