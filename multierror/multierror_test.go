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

func TestFormat(t *testing.T) {
	t.Parallel()

	t.Run("it format errors with the formatter function", func(t *testing.T) {
		e1 := errors.New("e1")
		e2 := errors.New("e2")
		e3 := errors.New("e3")

		formatter := func(es []error) string {
			s := ""
			for _, e := range es {
				s = s + e.Error()
			}
			return s
		}

		err := multierror.Append(e1, e2, e3)
		res := multierror.Format(err, formatter)
		assert.Equal(t, "e1e2e3", res)

		res = multierror.Format(errors.New("test error"), formatter)
		assert.Equal(t, "test error", res)

	})
}
