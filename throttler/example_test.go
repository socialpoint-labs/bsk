package throttler_test

import (
	"fmt"

	"github.com/socialpoint-labs/bsk/throttler"
	"golang.org/x/net/context"
	"sync/atomic"
	"time"
)

func ExampleThrottler_Throttle() {
	ctx := context.Background()
	th := throttler.NewThrottler(3, time.Second)
	th.Start(ctx)

	channelIn := make(chan string)
	var executions uint32

	action := &action{
		in:         channelIn,
		executions: &executions,
	}

	for i := 0; i < 5; i++ {
		err := th.Throttle(action)
		if err == nil {
			channelIn <- "hello"
		}
	}

	th.Stop()

	fmt.Print(fmt.Sprintf("Num executions: %d", atomic.LoadUint32(&executions)))

	// Output: Num executions: 3
}

type action struct {
	in         chan string
	executions *uint32
}

func (a *action) Execute() {
	<-a.in
	atomic.AddUint32(a.executions, 1)
}
