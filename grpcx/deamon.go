package grpcx

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"

	"github.com/socialpoint-labs/bsk/contextx"
)

// Daemon represents a runnable gRPC daemon
type Daemon struct {
	port int
	svr  *grpc.Server
	apps []Application
}

// Application is a contextx.Runner that can register gRPC services
type Application interface {
	contextx.Runner
	RegisterGRPC(*grpc.Server)
}

func NewDaemon(port int, svr *grpc.Server, apps ...Application) Daemon {
	return Daemon{
		port: port,
		svr:  svr,
		apps: apps,
	}
}

// Run registers and runs every application of the Daemon with a gRPC server,
// then starts it and gracefully shuts it down once the context is done.
func (d Daemon) Run(ctx context.Context) {
	for _, app := range d.apps {
		app.RegisterGRPC(d.svr)

		go app.Run(ctx)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", d.port))
	if err != nil {
		panic(err.Error())
	}

	defer d.svr.GracefulStop()

	go func() {
		if err := d.svr.Serve(listener); err != nil {
			panic(err.Error())
		}
	}()

	<-ctx.Done()
}
