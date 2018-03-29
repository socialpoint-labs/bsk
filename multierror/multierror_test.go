package multierror_test

import (
	"errors"
	"testing"

	"github.com/socialpoint-labs/bsk/multierror"
	"github.com/stretchr/testify/assert"
)

func TestAppend(t *testing.T) {
	t.Parallel()

	t.Run("it returns nil if no errors", func(t *testing.T) {
		assert.Nil(t, multierror.Append(multierror.Append()))
		assert.Nil(t, multierror.Append(nil))

		var e error
		assert.Nil(t, multierror.Append(e))
	})

	t.Run("it appends multiple errors", func(t *testing.T) {
		e1 := errors.New("e1")
		e2 := errors.New("e2")
		e3 := errors.New("e3")

		assert.Error(t, multierror.Append(e1, e2, e3))
		assert.Len(t, multierror.Append(e1, e2, e3), 3)
	})

	t.Run("it flattens multiple errors", func(t *testing.T) {
		e1 := errors.New("e1")
		e2 := errors.New("e2")
		e3 := errors.New("e3")
		me := multierror.Append(e1, e2, e3)

		assert.Len(t, multierror.Append(e1, e2, me, e3), 6)
	})

	t.Run("it does not append nil errors", func(t *testing.T) {
		e1 := errors.New("e1")
		e2 := errors.New("e2")
		e3 := errors.New("e3")

		assert.Len(t, multierror.Append(nil, e1, e2, nil, e3, nil), 3)
	})
}

func TestWalk(t *testing.T) {
	t.Parallel()

	e1 := errors.New("e1")
	e2 := errors.New("e2")
	e3 := errors.New("e3")

	var s string
	walker := func(i int, e error) { s = s + e.Error() }

	err := multierror.Append(e1, e2, e3)
	multierror.Walk(err, walker)
	assert.Equal(t, "e1e2e3", s)

	s = ""
	multierror.Walk(errors.New("test error"), walker)
	assert.Equal(t, "test error", s)
}
