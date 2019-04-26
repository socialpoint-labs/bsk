package dispatcher

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testInterface1 interface {
	TestFunc1()
}

type testInterface2 interface {
	TestFunc2()
}

type testEvent1 struct {
}

type testEvent2 struct {
	triggered bool
}

func (testEvent1) TestFunc1()  {}
func (*testEvent2) TestFunc2() {}

func TestStaticListener(t *testing.T) {
	assert := assert.New(t)

	dispatcher := New()

	triggered := false
	err := dispatcher.AddListener(func(testEvent1) { triggered = true })
	assert.NoError(err)

	err = dispatcher.AddListener(func(testEvent2) { t.Error("wrong listener type triggered") })
	assert.NoError(err)

	dispatcher.Dispatch(testEvent1{})
	assert.True(triggered, "static listener failed to trigger")
}

func TestPointerListener(t *testing.T) {
	assert := assert.New(t)

	testEvent := new(testEvent2)
	dispatcher := New()

	err := dispatcher.AddListener(func(ev *testEvent2) { ev.triggered = true })
	assert.NoError(err)

	err = dispatcher.AddListener(func(testEvent2) { t.Error("non-pointer listener triggered on pointer type") })
	assert.NoError(err)

	dispatcher.Dispatch(testEvent)
	assert.True(testEvent.triggered, "pointer listener failed to trigger")
}

func TestInterfaceListener(t *testing.T) {
	assert := assert.New(t)

	dispatcher := New()

	triggered := false
	err := dispatcher.AddListener(func(testInterface1) { triggered = true })
	assert.NoError(err)

	err = dispatcher.AddListener(func(testInterface2) { t.Error("interface listener triggerd on non-matching type") })
	assert.NoError(err)

	dispatcher.Dispatch(testEvent1{})
	assert.True(triggered, "interface listener failed to trigger")
}

func TestEmptyInterfaceListener(t *testing.T) {
	assert := assert.New(t)

	triggered := false
	dispatcher := New()

	err := dispatcher.AddListener(func(interface{}) { triggered = true })
	assert.NoError(err)

	dispatcher.Dispatch("this should match interface{}")
	assert.True(triggered)
}

func TestMultipleListeners(t *testing.T) {
	assert := assert.New(t)

	triggered1, triggered2 := false, false
	dispatcher := New()

	err := dispatcher.AddListener(func(testEvent1) { triggered1 = true })
	assert.NoError(err)

	err = dispatcher.AddListener(func(testEvent1) { triggered2 = true })
	assert.NoError(err)

	dispatcher.Dispatch(testEvent1{})

	assert.True(triggered1 && triggered2)
}

func TestVariadicListeners(t *testing.T) {
	assert := assert.New(t)

	triggered1, triggered2 := false, false
	dispatcher := New()

	listeners := []interface{}{
		func(testEvent1) { triggered1 = true },
		func(testEvent1) { triggered2 = true },
	}
	err := dispatcher.AddListener(listeners...)
	assert.NoError(err)

	dispatcher.Dispatch(testEvent1{})
	assert.True(triggered1 && triggered2)
}

func TestBadListenerWrongInputs(t *testing.T) {
	assert := assert.New(t)
	dispatcher := New()

	err := dispatcher.AddListener(func() {})
	assert.Error(err)
	assert.NotEmpty(err.Error())

}

func TestBadListenerWrongType(t *testing.T) {
	assert := assert.New(t)
	dispatcher := New()

	err := dispatcher.AddListener("this is not a function")
	assert.Error(err)
	assert.NotEmpty(err.Error())
}

func TestAsynchronousDispatch(t *testing.T) {
	assert := assert.New(t)

	triggered := make(chan bool)

	dispatcher := New()

	err := dispatcher.AddListener(func(testEvent1) { triggered <- true })
	assert.NoError(err)

	go dispatcher.Dispatch(testEvent1{})

	select {
	case <-triggered:
	case <-time.After(time.Second):
		assert.Fail("asynchronous dispatch failed to trigger listener")
	}
}

func TestDispatchPointerToValueInterfaceListener(t *testing.T) {
	assert := assert.New(t)
	dispatcher := New()

	triggered := false
	err := dispatcher.AddListener(func(ev testInterface1) {
		triggered = true
	})
	assert.NoError(err)

	dispatcher.Dispatch(&testEvent1{})
	assert.True(triggered, "Dispatch by pointer failed to trigger interface listener")
}

func TestDispatchValueToValueInterfaceListener(t *testing.T) {
	assert := assert.New(t)
	dispatcher := New()

	triggered := false

	err := dispatcher.AddListener(func(ev testInterface1) {
		triggered = true
	})
	assert.NoError(err)

	dispatcher.Dispatch(testEvent1{})
	assert.True(triggered, "Dispatch by value failed to trigger interface listener")
}

func TestDispatchPointerToPointerInterfaceListener(t *testing.T) {
	assert := assert.New(t)
	triggered := false
	dispatcher := New()

	err := dispatcher.AddListener(func(testInterface2) { triggered = true })
	assert.NoError(err)

	dispatcher.Dispatch(&testEvent2{})

	assert.True(triggered, "interface listener failed to trigger for pointer")
}

func TestDispatchValueToPointerInterfaceListener(t *testing.T) {
	assert := assert.New(t)
	dispatcher := New()

	err := dispatcher.AddListener(func(testInterface2) {
		assert.Fail("interface listener triggered for value dispatch")
	})
	assert.NoError(err)

	dispatcher.Dispatch(testEvent2{})
}
