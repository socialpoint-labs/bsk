package grpcx

import (
	"context"
	"fmt"
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
// Either a port or a listener is needed for the daemon to run. If no listener is given,
// a TCP listener will be created for the given port.
//
// Typically, you'll want to pass one or more Applications. The daemon will register the gRPC
// endpoints of every application and run it with the runner context.
// A daemon can run without any Applications. This can be interesting for certain use cases where you want
// to manage the server by yourself and delegate the lifecycle management.
//
// Errors that might happen during the listener setup and running the gRPC server will be passed
// through an optional error func. If none is given, they'll be printed to the standard output.
type Daemon struct {
	svr  *grpc.Server
	port int
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
	port int
	lis  net.Listener
	apps []Application
	ef   func(error)
}

func WithPort(p int) DaemonOption {
	return func(opts *daemonOptions) {
		opts.port = p
	}
}

func WithListener(lis net.Listener) DaemonOption {
	return func(opts *daemonOptions) {
		opts.lis = lis
	}
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

// NewDaemon creates a gRPC Daemon for the given listener.
func NewDaemon(svr *grpc.Server, opts ...DaemonOption) Daemon {
	options := &daemonOptions{
		ef: defaultErrorFunc,
	}
	for _, opt := range opts {
		opt(options)
	}

	return Daemon{
		svr:  svr,
		port: options.port,
		lis:  options.lis,
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

func defaultErrorFunc(err error) {
	log.Println(err.Error())
}
