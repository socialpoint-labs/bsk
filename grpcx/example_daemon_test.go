package grpcx_test

import (
	"context"
	"errors"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/test/bufconn"

	"github.com/socialpoint-labs/bsk/grpcx"
)

func ExampleDaemon_Run() {
	server := exampleServer()
	lis := bufconn.Listen(1024 * 1024)
	defer lis.Close()

	ctx, cancel := context.WithCancel(context.Background())
	app := exampleApplication{}
	dae := grpcx.NewDaemon(server, grpcx.WithListener(lis), grpcx.WithApplications(app))
	go dae.Run(ctx)

	exampleCall(ctx, lis)

	cancel()

	// Output: example rpcs registered
	// example application run
	// /test.service/test.call rpc handled!
	// got response from server: bye
}

func exampleServer() *grpc.Server {
	encoding.RegisterCodec(noopCodec{})
	return grpc.NewServer(grpc.UnknownServiceHandler(exampleServiceHandler))
}

func exampleServiceHandler(srv interface{}, stream grpc.ServerStream) error {
	fullMethodName, ok := grpc.MethodFromServerStream(stream)
	if !ok {
		return errors.New("could not determine method from server stream")
	}

	fmt.Println(fullMethodName, "rpc handled!")

	return stream.SendMsg([]byte("bye"))
}

func exampleCall(ctx context.Context, lis *bufconn.Listener) {
	cc, err := grpc.DialContext(
		ctx, "",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.CustomCodecCallOption{Codec: noopCodec{}}),
	)
	if err != nil {
		panic(err)
	}
	defer cc.Close()

	var out []byte
	err = cc.Invoke(ctx, "/test.service/test.call", nil, &out)
	if err != nil {
		panic(err)
	}

	fmt.Println("got response from server:", string(out))
}

type exampleApplication struct {
}

func (e exampleApplication) Run(ctx context.Context) {
	fmt.Println("example application run")
}

func (e exampleApplication) RegisterGRPC(registrar grpc.ServiceRegistrar) {
	fmt.Println("example rpcs registered")
}
