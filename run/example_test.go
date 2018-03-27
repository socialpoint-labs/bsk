package run_test

import (
	"context"
	"fmt"
	"time"

	"github.com/socialpoint-labs/bsk/run"
)

func ExampleWithRetry() {
	err := run.WithRetry(
		run.ConstantBackoff(3, time.Millisecond),

		func(e error) run.Result {
			fmt.Println("Retrying...")
			return run.Retry
		},

		func(i context.Context) error {
			return nil
		},
	)(context.Background())

	if err != nil {
		panic(err)
	}

	// Output:
	// Retrying...
	// Retrying...
	// Retrying...
}

func ExampleWithDeadline() {
	err := run.WithDeadline(time.Now().Add(time.Millisecond*10), func(ctx context.Context) error {
		<-ctx.Done()
		fmt.Println("Deadline...")

		return nil
	})(context.Background())

	if err != nil {
		panic(err)
	}

	// Output:
	// Deadline...
}
