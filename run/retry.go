package run

import (
	"context"
	"math/rand"
	"time"
)

// WithRetry enables an application to handle transient failures by transparently retrying a failed operation.
func WithRetry(backoffs []time.Duration, classifier func(error) Result, fn func(context.Context) error) func(context.Context) error {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	return func(ctx context.Context) error {
		var err error
		for _, backoff := range backoffs {
			err = fn(ctx)

			switch classifier(err) {
			case Succeed, Cancel:
				return err

			case Retry:
				jitter := rnd.Float64() / 2
				sleep := backoff + time.Duration(jitter*float64(backoff))
				Sleep(ctx, sleep)
			}
		}

		return err
	}
}

// Result is the type returned by error classifier functions to indicate whether a retry should proceed.
type Result int

// In case of errors it can handle the failure using the following strategies
const (
	// Succeed indicates that the run was as a success, there is no need to retry
	Succeed Result = iota

	// Cancel indicates a hard failure that should not be retried.
	// It indicates that the failure isn't transient or is unlikely to be successful if repeated,
	// the application should cancel the operation and report the error.
	// For example, an authentication failure caused by providing invalid credentials is not
	// likely to succeed no matter how many times it's attempted.
	Cancel

	// Retry indicates a soft failure and should be retried
	// If the specific fault reported is unusual or rare, it might have been caused by unusual circumstances
	// such as a network packet becoming corrupted while it was being transmitted. In this case,
	// the application could retry the failing request again immediately because the same failure is unlikely
	// to be repeated and the request will probably be successful.
	Retry
)

// NotNilClassifier is an error classifier function that returns Succeed if error is nil,
// otherwise it returns Retry
func NotNilClassifier() func(error) Result {
	return func(err error) Result {
		if err == nil {
			return Succeed
		}

		return Retry
	}
}

// ExponentialBackoff generates an exponential back-off strategy, retrying the given
// number of times and doubling the waiting time in every retry.
func ExponentialBackoff(retries int, initial time.Duration) []time.Duration {
	durations := make([]time.Duration, retries)
	for i := range durations {
		durations[i] = initial
		initial *= 2
	}

	return durations
}

// ConstantBackoff generates a back-off strategy of retrying the given
// number of times and waiting the specified time duration after each one.
func ConstantBackoff(retries int, backoff time.Duration) []time.Duration {
	durations := make([]time.Duration, retries)
	for i := range durations {
		durations[i] = backoff
	}

	return durations
}
