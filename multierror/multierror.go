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

// Format allows to format the string representation of a multi error with an alternative formatter function
func Format(err error, formatter func([]error) string) string {
	switch err := err.(type) {
	case errors:
		return formatter(err)

	default:
		return formatter([]error{err})
	}
}

type errors []error

func (es errors) Error() string {
	items := make([]string, len(es))
	for i, err := range es {
		items[i] = fmt.Sprintf("- %s", err)
	}

	return fmt.Sprintf("%d errors occurred:\n%s", len(es), strings.Join(items, "\n"))
}
