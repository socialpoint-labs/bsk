package throttler

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

// ThrottledAction is a common interface for actions that can be executed
type ThrottledAction interface {
	Execute()
}

// Throttler is a structure that specifies a maximum number of actions to be executed in parallel
type Throttler struct {
	currentExecutions uint32
	maxExecutions     uint32
	ticker            *time.Ticker
}

// NewThrottler returns the reference to a new Instace of a Throttler struct, the argument max will control how many
// actions can be executed per time duration using this instance of Throttler
func NewThrottler(max int, duration time.Duration) *Throttler {
	ticker := time.NewTicker(duration)

	return &Throttler{
		maxExecutions: uint32(max),
		ticker:        ticker,
	}
}

// Start the throttler's ticker
func (t *Throttler) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-t.ticker.C:
				t.resetCounter()
			case <-ctx.Done():
				t.Stop()
				return
			}
		}
	}()
}

// Stop the throttler's ticker
func (t *Throttler) Stop() {
	t.ticker.Stop()
}

// Throttle receives an action which we want to execute
// There is a maximum (max) number of methods to be executed per time interval, specified in the Throttler struct
// - if more than max methods are provided, only max will be executed, the rest will return error
// - it is responsibility of the caller to retry the discarded actions if needed
// - if the action to execute runs in a goroutine, at any given moment in time, the max actions to execute may be exceeded
func (t *Throttler) Throttle(action ThrottledAction) error {
	if !t.executionAllowed() {
		return fmt.Errorf("Maximum number of executions reached: %d", t.maxExecutions)
	}

	atomic.AddUint32(&t.currentExecutions, 1)
	go t.execute(action)
	return nil
}

func (t *Throttler) executionAllowed() bool {
	return atomic.LoadUint32(&t.currentExecutions) < t.maxExecutions
}

func (t *Throttler) execute(action ThrottledAction) {
	action.Execute()
}

func (t *Throttler) resetCounter() {
	atomic.StoreUint32(&t.currentExecutions, 0)
}
