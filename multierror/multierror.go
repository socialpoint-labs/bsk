package multierror

import (
	"fmt"
	"strings"
)

// Append appends error elements to the end of an error slice, returning a new error
func Append(errs ...error) error {
	err := errors{}

	for _, e := range errs {
		switch e := e.(type) {
		case errors:
			if e != nil {
				err = append(err, e...)
			}
		default:
			if e != nil {
				err = append(err, e)
			}
		}
	}

	if len(err) == 0 {
		return nil
	}

	return err
}

// Walk allows to traverse a multi error with an custom function
func Walk(e error, f func(int, error)) {
	var es errors

	switch e := e.(type) {
	case errors:
		es = e
	default:
		es = []error{e}
	}

	for i, e := range es {
		f(i, e)
	}
}

type errors []error

func (es errors) Error() string {
	items := make([]string, len(es))
	Walk(es, func(i int, e error) { items[i] = fmt.Sprintf("- %s", e) })

	return fmt.Sprintf("%d errors occurred:\n%s", len(es), strings.Join(items, "\n"))
}
