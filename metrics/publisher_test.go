package metrics

import (
	"testing"
	"time"

	"context"

	"bufio"
	"net"

	"github.com/socialpoint-labs/bsk/netutil"
	"github.com/stretchr/testify/assert"
)

func TestPublisherImplementsMetrics(t *testing.T) {
	check := func(m Metrics) {}
	check(&Publisher{})
}

func TestPublisherForwardMetrics(t *testing.T) {
	assert := assert.New(t)
	rec := make(recorder)

	publisher := NewPublisher(rec, StatsDEncoder, FlushEvery3s, nil)
	go publisher.Run(context.Background())

	counter := publisher.Counter("commands_executed", Tag{"host", "life"}, Tag{"project", "bsk"})
	counter.Add(1)
	counter.WithTags(NewTag("cfoo", "cbar")).Inc()

	gauge := publisher.Gauge("memory", Tag{"host", "life"}, Tag{"project", "bsk"})
	gauge.WithTags(NewTag("gfoo", "gbar")).Update(100)

	publisher.Flush()

	expected := "commands_executed:1|c|@1.0000|#host:life,project:bsk\ncommands_executed:1|c|@1.0000|#host:life,project:bsk,cfoo:cbar\nmemory:100|g|@1.0000|#host:life,project:bsk,gfoo:gbar\n"

	assert.Equal(expected, <-rec)
}

func TestPublisherFlushBufferWhenMaxSizeIsExceeded(t *testing.T) {
	rec := make(recorder, 1024)
	assert := assert.New(t)
	timeout := time.After(time.Second * 3)

	// Create a publisher that takes a long time to flush
	publisher := NewPublisher(rec, StatsDEncoder, time.Hour, nil)
	go publisher.Run(context.Background())
	counter := publisher.Counter("commands_executed", Tag{"host", "life"}, Tag{"project", "bsk"})

	// Increment the counter infinitely.
	go func() {
		for {
			counter.Inc()
		}
	}()

	// This tests and proof the correctness of  the invariant that a
	// publisher must flush when certain size of the buffer is exceeded,
	// even before the flush time
	select {
	case <-rec:
		// If we received something in the recorder, this means that the publisher flushed the buffer
		// Everything as expected then!
		return
	case <-timeout:
		// We reached the timeout and no flush occurred, bad news!
		assert.Fail("timeout reached and the publisher did't flush out the metrics")
		return
	default:
	}
}

func TestPublisherFlushMetricsToRealUDPServer(t *testing.T) {
	assert := assert.New(t)

	addr := netutil.FreeUDPAddr()

	server, err := net.ListenUDP("udp", addr)
	assert.NoError(err)

	client, err := net.DialUDP("udp", nil, addr)
	assert.NoError(err)

	publisher := NewPublisher(client, StatsDEncoder, time.Millisecond, nil)
	go publisher.Run(context.Background())
	counter := publisher.Counter("test")

	counter.Add(123)

	reader := bufio.NewReader(server)

	line, err := reader.ReadString('\n')

	assert.NoError(err)
	assert.Equal("test:123|c|@1.0000|#\n", line)
}

func TestTimerEvent(t *testing.T) {
	assert := assert.New(t)

	addr := netutil.FreeUDPAddr()

	server, err := net.ListenUDP("udp", addr)
	assert.NoError(err)

	client, err := net.DialUDP("udp", nil, addr)
	assert.NoError(err)

	publisher := NewPublisher(client, StatsDEncoder, time.Millisecond, nil)
	go publisher.Run(context.Background())

	timer := publisher.Timer("test")

	timer.Start()
	timer.Stop()

	reader := bufio.NewReader(server)

	line, err := reader.ReadString('\n')

	assert.NoError(err)
	assert.Contains(line, "test:")
	assert.Contains(line, "|ms|@1.0000|#\n")
}

type recorder chan string

func (r recorder) Write(b []byte) (n int, err error) {
	r <- string(b)
	return len(b), nil
}
