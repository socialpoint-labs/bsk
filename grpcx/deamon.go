package grpcx

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/socialpoint-labs/bsk/contextx"
)

// Daemon represents a runnable gRPC daemon.
//
// A gRPC Daemon will manage the lifecycle of the given gRPC server, gracefully shutting it down
// once the runner context is done.
//
// Typically, you'll want to pass one or more Applications as options. The daemon will register the gRPC
// endpoints of every application and run it with the runner context.
// A daemon can run without any Applications. This can be interesting for certain use cases where you want
// to manage the server by yourself and delegate the lifecycle management.
//
// Errors that might happen  running the gRPC server will be passed through an optional error func.
// If none is given, they'll be printed to the standard output.
type Daemon struct {
	svr  *grpc.Server
	lis  net.Listener
	apps []Application
	ef   func(error)
}

// Application is a contextx.Runner that can register gRPC services
type Application interface {
	contextx.Runner
	RegisterGRPC(grpc.ServiceRegistrar)
}

type DaemonOption func(*daemonOptions)

type daemonOptions struct {
	apps []Application
	ef   func(error)
}

func WithApplications(apps ...Application) DaemonOption {
	return func(opts *daemonOptions) {
		opts.apps = apps
	}
}

func WithErrorFunc(ef func(error)) DaemonOption {
	return func(opts *daemonOptions) {
		opts.ef = ef
	}
}

// NewDaemon creates a gRPC Daemon for the given gRPC server and listener.
func NewDaemon(svr *grpc.Server, lis net.Listener, opts ...DaemonOption) Daemon {
	options := &daemonOptions{
		ef: defaultErrorFunc,
	}
	for _, opt := range opts {
		opt(options)
	}

	return Daemon{
		svr:  svr,
		lis:  lis,
		apps: options.apps,
		ef:   options.ef,
	}
}

// Run registers every application of the Daemon with a gRPC server and runs them,
// then starts the server and gracefully shuts it down once the context is done.
func (d Daemon) Run(ctx context.Context) {
	for _, app := range d.apps {
		app.RegisterGRPC(d.svr)

		go app.Run(ctx)
	}

	defer d.svr.GracefulStop()

	go func() {
		if err := d.svr.Serve(d.lis); err != nil {
			d.ef(err)
			return
		}
	}()

	<-ctx.Done()
}

func defaultErrorFunc(err error) {
	log.Println(err.Error())
}
