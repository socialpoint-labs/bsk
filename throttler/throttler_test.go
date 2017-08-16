package throttler

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

var (
	mux           sync.Mutex
	numExecutions uint32
)

//Test that we can execute the number of actions allowed by the scheduler
func TestExecuteActions(t *testing.T) {
	numExecutions = 0
	channelIn := make(chan string)
	maxProc := 10
	assert := assert.New(t)

	th := NewThrottler(maxProc, time.Second*2)
	th.Start(context.Background())

	for i := 0; i < maxProc; i++ {
		a := &action{in: channelIn}
		err := th.Throttle(a)
		assert.Nil(err)
	}

	assert.Equal(th.currentExecutions, uint32(maxProc))

	for i := 0; i < maxProc; i++ {
		channelIn <- "hello"
	}

	assert.Equal(uint32(maxProc), numExecutions)

	th.Stop()
}

//Test that any actions executed over the limit allowed, will not be executed by the throttler
func TestThrottleActions(t *testing.T) {
	numExecutions = 0
	channelIn := make(chan string)
	maxProc := 10
	discardedJobs := 0

	th := NewThrottler(maxProc, time.Second*2)
	th.Start(context.Background())

	for i := 0; i < maxProc*2; i++ {
		a := &action{in: channelIn}
		err := th.Throttle(a)
		if err != nil {
			discardedJobs++
		}
	}

	for i := 0; i < maxProc; i++ {
		channelIn <- "hello"
	}

	assert := assert.New(t)
	assert.Equal(uint32(maxProc), numExecutions)

	th.Stop()
}

//Test that the limit is only applied based on the interval, not on the amount of running actions
func TestThrottleIsBasedOnIntervals(t *testing.T) {
	numExecutions = 0
	channelIn := make(chan string)
	maxProc := 1
	assert := assert.New(t)
	interval := time.Millisecond * 5

	th := NewThrottler(maxProc, interval)
	th.Start(context.Background())

	a := &action{in: channelIn}
	err := th.Throttle(a)
	assert.Nil(err)

	time.Sleep(interval * 5)

	err = th.Throttle(a)

	assert.Nil(err)

	channelIn <- "hello"
	channelIn <- "hello"

	assert.Equal(uint32(maxProc*2), uint32(numExecutions))

}

//Test that the throttler is concurrent safe
func TestThrottleConcurrency(t *testing.T) {
	numExecutions = 0
	channelIn := make(chan string)
	maxProc := 8
	assert := assert.New(t)
	interval := time.Second

	th := NewThrottler(maxProc, interval)
	th.Start(context.Background())

	a := &action{in: channelIn}
	for i := 0; i < 5; i++ {
		go throttleActions(th, a, channelIn)
	}

	for i := 0; i < maxProc; i++ {
		channelIn <- "hello"
	}

	assert.Equal(uint32(maxProc), uint32(numExecutions))
}

type action struct {
	in chan string
}

func (a *action) Execute() {
	mux.Lock()
	numExecutions++
	mux.Unlock()
	<-a.in
}

func throttleActions(th *Throttler, action *action, channel chan string) {
	for i := 0; i < 5; i++ {
		_ = th.Throttle(action)
	}
}
