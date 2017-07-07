package httpx

import (
	"net/http"

	"strings"

	"context"
)

// NewRouter returns a new Router struct
func NewRouter(opts ...Option) *Router {
	options := &options{}
	for _, o := range opts {
		o(options)
	}

	if options.responder == nil {
		options.responder = NewResponder()
	}

	return &Router{
		options: options,
		mux:     http.NewServeMux(),
	}
}

// Router allows to route requests to according handlers
type Router struct {
	options     *options
	routes      []*route
	mux         *http.ServeMux
	initialized bool
}

// Route registers a route pattern against its handler
func (router *Router) Route(pattern string, handler http.Handler, decorators ...Decorator) {
	router.routes = append(router.routes, &route{pattern, handler, decorators})
}

// ServeHTTP implements http.Handler
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !router.initialized {
		registerRoutes(router.mux, router, "", []Decorator{})
		router.initialized = true
	}

	h, _ := router.mux.Handler(r)

	ctx := context.WithValue(r.Context(), responderKey, router.options.responder)

	r = r.WithContext(ctx)

	h.ServeHTTP(w, r)
}

type route struct {
	pattern    string
	handler    http.Handler
	decorators []Decorator
}

func registerRoutes(mux *http.ServeMux, router *Router, prefix string, decorators []Decorator) {
	for _, r := range router.routes {
		uri := prefix + r.pattern
		if uri != "/" {
			uri = strings.TrimRight(uri, "/")
		}

		if child, ok := r.handler.(*Router); ok {
			registerRoutes(mux, child, uri, append(r.decorators, decorators...))
		} else {
			if uri == "/" {
				r.decorators = append(r.decorators, RootDecorator())
			}

			if prefix != "" {
				r.decorators = append(r.decorators, StripPrefixDecorator(prefix))
			}

			for _, decorator := range append(r.decorators, decorators...) {
				r.handler = decorator(r.handler)
			}

			mux.Handle(uri, r.handler)
		}
	}
}

// Option is the common type of functions that set options
type Option func(*options)

type options struct {
	responder *Responder
}

type contextKey int

const (
	responderKey contextKey = iota
)
