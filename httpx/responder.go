package httpx

import (
	"net/http"
)

// NewResponder returns a new Responder
func NewResponder() *Responder {
	return &Responder{}
}

// Responder actually write responses, typically at the end of an HTTP request
type Responder struct {
	// OnErr is a function that gets called when an error occurs while responding.
	// By default, the error panic but you may
	// use Options.OnErrLog to just log the error out instead,
	// or provide your own.
	OnErr func(err error)

	// Encoder is a function field that gets the encoder to
	// use to respond to the specified http.Request.
	// If nil, JSON will be used.
	Encoder func(w http.ResponseWriter, r *http.Request) Encoder

	// Before is called for before each response is written
	// and gives user code the chance to mutate the status or data.
	// Useful for handling different types of data differently (like errors),
	// enveloping the response, setting common HTTP headers etc.
	Before func(w http.ResponseWriter, r *http.Request, status int, data interface{}) (int, interface{})

	// After is called after each response.
	// Useful for logging activity after a response has been written.
	After func(w http.ResponseWriter, r *http.Request, status int, data interface{})

	// StatusData is a function field that gets the data to respond with when
	// WithStatus is called.
	// By default, the function will return an object that looks like this:
	//     {"status":"Not Found","code":404}
	StatusData func(w http.ResponseWriter, r *http.Request, status int) interface{}
}

// Respond uses the http.ResponseWriter for writing the response data and status
func (o *Responder) Respond(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	encoder := JSONEncoder // JSON by default

	if o.Before != nil {
		status, data = o.Before(w, r, status, data)
	}

	if o.Encoder != nil {
		encoder = o.Encoder(w, r)
	}

	// Actually write the response
	w.Header().Set("Content-Type", encoder.ContentType(w, r))

	w.WriteHeader(status)

	if err := encoder.Encode(w, r, data); err != nil {
		if o.OnErr != nil {
			o.OnErr(err)
		} else {
			panic("respond: " + err.Error())
		}
	}

	if o.After != nil {
		o.After(w, r, status, data)
	}
}

// WithStatus responds to the client with the specified status.
// Responder.StatusData will be called to obtain the data payload, or a default
// payload will be returned:
//     {"status":"I'm a teapot","code":418}
func (o *Responder) WithStatus(w http.ResponseWriter, r *http.Request, status int) {
	var data interface{}
	if o.StatusData != nil {
		data = o.StatusData(w, r, status)
	} else {
		data = map[string]interface{}{"status": http.StatusText(status), "code": status}
	}

	o.Respond(w, r, status, data)
}

// RespondWith returns an option that sets a responder
func RespondWith(r *Responder) Option {
	return func(o *options) {
		o.responder = r
	}
}

// Respond extracts a responder from the context and writes the data and status to the http.ResponseWriter
func Respond(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	var responder *Responder
	var ok bool
	if responder, ok = r.Context().Value(responderKey).(*Responder); !ok {
		responder = NewResponder()
	}

	responder.Respond(w, r, status, data)
}
