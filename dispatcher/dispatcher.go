package dispatcher

import (
	"fmt"
	"reflect"
	"sync"
)

// BadListenerError is raised when AddListener is called with an invalid listener function.
type BadListenerError string

func (err BadListenerError) Error() string {
	return fmt.Sprintf("Bad listener func: %s", string(err))
}

// New returns a new dispatcher
func New() *Dispatcher {
	return &Dispatcher{
		listeners:  make(map[reflect.Type][]interface{}),
		interfaces: make([]reflect.Type, 0),
	}
}

// A Dispatcher sends events to the listeners that listen to the
// events of a certain type.
type Dispatcher struct {
	lock       sync.RWMutex
	listeners  map[reflect.Type][]interface{}
	interfaces []reflect.Type
}

// AddListener registers a listener function that will be called when a matching
// event is dispatched. The type of the function's first (and only) argument
// declares the event type (or interface) to listen for.
func (d *Dispatcher) AddListener(listeners ...interface{}) error {
	// check for errors
	for _, listener := range listeners {
		listenerType := reflect.TypeOf(listener)
		if listenerType.Kind() != reflect.Func || listenerType.NumIn() != 1 {
			return BadListenerError("listener must be a function that takes exactly one argument")
		}
	}

	// store them
	d.lock.Lock()
	defer d.lock.Unlock()
	for _, listener := range listeners {
		listenerType := reflect.TypeOf(listener)
		// the first input parameter is the event
		eventType := listenerType.In(0)

		// keep a list of listeners for each event type
		d.listeners[eventType] = append(d.listeners[eventType], listener)

		// if the listener is an interface store it in a separate list
		// so we can check non-interface objects against all interfaces
		if eventType.Kind() == reflect.Interface {
			d.interfaces = append(d.interfaces, eventType)
		}
	}

	return nil
}

// Dispatch sends an event to all registered listeners that were declared
// to accept values of the event's type, or interfaces that the value implements.
func (d *Dispatcher) Dispatch(ev interface{}) {
	d.lock.RLock()
	defer d.lock.RUnlock()

	evType := reflect.TypeOf(ev)
	vals := []reflect.Value{reflect.ValueOf(ev)}

	// Call listeners for the actual static type
	d.callListeners(evType, vals)

	// Also check if the type implements any of the registered interfaces
	for _, in := range d.interfaces {
		if evType.Implements(in) {
			d.callListeners(in, vals)
		}
	}
}

func (d *Dispatcher) callListeners(t reflect.Type, vals []reflect.Value) {
	for _, fn := range d.listeners[t] {
		reflect.ValueOf(fn).Call(vals)
	}
}
