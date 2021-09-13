package timex_test

import (
	"context"
	"testing"
	"time"

	"github.com/socialpoint-labs/bsk/timex"
	"github.com/stretchr/testify/assert"
)

func TestRunInterval(t *testing.T) {
	assert := assert.New(t)

	ch := make(chan time.Time)

	f := func() {
		ch <- time.Now()
	}

	interval := 10 * time.Millisecond
	runner := timex.IntervalRunner(interval, f)

	ctx, cancel := context.WithCancel(context.Background())
	go runner.Run(ctx)

	t1 := <-ch
	t2 := <-ch
	t3 := <-ch

	assert.True(t1.Before(t2), "t1 should be before t2")
	assert.True(t2.Before(t3), "t2 should be before t3")

	assert.WithinDuration(t2, t1, 2*interval)
	assert.WithinDuration(t3, t2, 2*interval)

	cancel()
}
