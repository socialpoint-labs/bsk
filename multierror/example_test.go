package multierror_test

import (
	"errors"
	"fmt"

	"github.com/socialpoint-labs/bsk/multierror"
)

func ExampleAppend() {
	var err error

	err = multierror.Append(err, errors.New("error 1"))
	err = multierror.Append(err, errors.New("error 2"))
	err = multierror.Append(err, errors.New("error 3"))
	err = multierror.Append(err, errors.New("error 4"), errors.New("error 5"))

	fmt.Println(err)

	// Output:
	// 5 errors occurred:
	// - error 1
	// - error 2
	// - error 3
	// - error 4
	// - error 5
}
