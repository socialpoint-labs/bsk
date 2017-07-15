package timex_test

import (
	"context"
	"errors"
	"time"

	"github.com/socialpoint-labs/bsk/timex"
)

func ExampleParse() {
	_, err := timex.Parse("")
	if err != nil {
		return
	}

	_, err = timex.Parse("2016-04-23 12:56")
	if err != nil {
		return
	}

	_, err = timex.Parse("-10 days")
	if err != nil {
		return
	}

	_, err = timex.Parse("-5 hours")
	if err != nil {
		return
	}

	_, err = timex.Parse("1464876005")
	if err != nil {
		return
	}

	// Output:
}

func ExampleParseFromDate() {
	_, err := timex.ParseFromDate("2016-04-23 12:56")
	if err != nil {
		return
	}

	// Output:
}

func ExampleParseFromDaysAgo() {
	_, err := timex.ParseFromDaysAgo("-1 day")
	if err != nil {
		return
	}

	// Output:
}

func ExampleParseFromHoursAgo() {
	_, err := timex.ParseFromHoursAgo("-1 hour")
	if err != nil {
		return
	}

	// Output:
}

func ExampleParseFromTimestamp() {
	_, err := timex.ParseFromTimestamp("1464876005")
	if err != nil {
		return
	}

	// Output:
}

func ExampleIntervalRunner() {
	f := func() {}

	runner := timex.IntervalRunner(time.Millisecond, f)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go runner.Run(ctx)

	// Output:
}

func ExampleIntervalRunner_stop_on_error() {
	alwaysFail := func() error {
		return errors.New("arbitrary error")
	}

	ctx, cancel := context.WithCancel(context.Background())
	f := func() {
		err := alwaysFail()
		if err != nil {
			cancel()
		}
	}

	runner := timex.IntervalRunner(time.Millisecond, f)

	go runner.Run(ctx)

	// Output:
}
