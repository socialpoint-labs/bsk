# package dispatcher

`package dispatcher` provides a reflect-based framework for dispatching events.

Events are values of any arbitrary type.

For example, any package can define an event type:

	package mypackage

	type MyEvent struct {
		field1, field2 string
	}

Then, any other package (e.g. a plugin) can listen for those events:

	package myplugin

	import (
		"event"
		"mypackage"
	)

	func onMyEvent(ev mypackage.MyEvent) {
		// do something with ev
	}

	func init() {
		event.AddListener(onMyEvent)
	}

Any registered listeners that accept a single argument of type MyEvent will
be called when a value of type MyEvent is dispatched:

	package myotherpackage

	import (
		"event"
		"mypackage"
	)

	func DoSomething() {
		ev := mypackage.MyEvent{
			field1: "foo",
			field2: "bar",
		}

		event.Dispatch(ev)
	}

In addition, listener functions that accept an interface type will be called
for any dispatched value that implements the specified interface.

A listener that accepts interface{} will be called for every event type.

Listeners can also accept pointer types, but they will only be called if the dispatch
site calls Dispatch() on a pointer.