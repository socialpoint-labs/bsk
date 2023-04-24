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
	lis  net.Listener
	svr  *grpc.Server
	ef   func(error)
	apps []Application
}

// Application is a contextx.Runner that can register gRPC services
type Application interface {
	contextx.Runner
	RegisterGRPC(grpc.ServiceRegistrar)
}

// NewDaemonWithPort creates a gRPC Daemon listening at the given port.
func NewDaemonWithPort(port int, svr *grpc.Server, ef func(error), apps ...Application) Daemon {
	return Daemon{
		port: port,
		svr:  svr,
		ef:   ef,
		apps: apps,
	}
}

// NewDaemon creates a gRPC Daemon for the given listener.
//
// The lifecycle of the listener will *not* be managed by this function, nor by the Daemon. The application
// needs to take manage it.
func NewDaemon(lis net.Listener, svr *grpc.Server, ef func(error), apps ...Application) Daemon {
	return Daemon{
		lis:  lis,
		svr:  svr,
		ef:   ef,
		apps: apps,
	}
}

// Run registers every application of the Daemon with a gRPC server and runs them,
// then starts the server and gracefully shuts it down once the context is done.
func (d Daemon) Run(ctx context.Context) {
	for _, app := range d.apps {
		app.RegisterGRPC(d.svr)

		go app.Run(ctx)
	}

	listener, err := d.listener()
	if err != nil {
		d.ef(err)
		return
	}

	defer d.svr.GracefulStop()

	go func() {
		if err := d.svr.Serve(listener); err != nil {
			d.ef(err)
			return
		}
	}()

	<-ctx.Done()
}

func (d Daemon) listener() (net.Listener, error) {
	if d.lis != nil {
		return d.lis, nil
	}

	return net.Listen("tcp", fmt.Sprintf(":%d", d.port))
}
