package grpcx_test

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/test/bufconn"

	"github.com/socialpoint-labs/bsk/grpcx"
)

func TestDaemon_Run_CancellingContext(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	callback := make(chan func())
	server := testServer(callback)
	lis := bufconn.Listen(1024 * 1024)
	defer lis.Close()

	ctx, cancel := context.WithCancel(context.Background())
	dae := grpcx.NewDaemon(server, grpcx.WithListener(lis), grpcx.WithErrorFunc(func(err error) { t.Fatal(err) }))
	go dae.Run(ctx)
	cli := newTestClient(lis)

	// regular call
	cli.call()
	callback <- func() {}
	resp1 := <-cli.resp

	// cancel context in the middle of a call
	cli.call()
	callback <- cancel
	resp2 := <-cli.resp

	// call after context is cancelled
	cli.call()
	err := <-cli.errs

	a.Equal("call #1", resp1)
	a.Equal("call #2", resp2)
	a.EqualError(err, `rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing closed"`)
}

func TestDaemon_Run_WithInvalidPort(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server := exampleServer()
	chErr := make(chan error)
	dae := grpcx.NewDaemon(server, grpcx.WithPort(-1), grpcx.WithErrorFunc(func(err error) {
		chErr <- err
	}))

	go dae.Run(ctx)

	a.EqualError(<-chErr, "listen tcp: address -1: invalid port")
}

func testServer(callback chan func()) *grpc.Server {
	var calls int

	encoding.RegisterCodec(noopCodec{})
	return grpc.NewServer(grpc.UnknownServiceHandler(func(srv interface{}, stream grpc.ServerStream) error {
		cb := <-callback
		cb()

		calls++
		return stream.SendMsg([]byte(fmt.Sprintf("call #%d", calls)))
	}))
}

type noopCodec struct{}

func (noopCodec) Marshal(v interface{}) ([]byte, error) {
	return v.([]byte), nil
}

func (noopCodec) Unmarshal(data []byte, v interface{}) error {
	*(v.(*[]byte)) = data
	return nil
}

func (noopCodec) Name() string {
	return "noop-codec"
}

func (noopCodec) String() string {
	return "noop-codec"
}

type testClient struct {
	lis  *bufconn.Listener
	resp chan string
	errs chan error
}

func newTestClient(lis *bufconn.Listener) *testClient {
	return &testClient{
		lis:  lis,
		resp: make(chan string),
		errs: make(chan error),
	}
}

func (c testClient) call() {
	go c.callAsync()
}

func (c testClient) callAsync() {
	ctx := context.Background()
	cc, err := grpc.DialContext(
		ctx, "",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return c.lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.CustomCodecCallOption{Codec: noopCodec{}}),
	)
	if err != nil {
		c.errs <- err
	}
	defer cc.Close()

	var out []byte
	err = cc.Invoke(ctx, "/test.service/test.call", nil, &out)
	if err != nil {
		c.errs <- err
	}

	c.resp <- string(out)
}
