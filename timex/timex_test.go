package timex_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/socialpoint-labs/bsk/timex"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		dateStr  string
		valid    bool
		expected string
	}{
		{"2016-04-23 12:56", true, "2016-04-23T12:56:00"},
		{"2016-04-23", true, "2016-04-23T00:00:00"},
		{"-1 day", true, now.Add(-1 * 24 * time.Hour).Format(time.RFC3339)},
		{"-20 days", true, now.Add(-20 * 24 * time.Hour).Format(time.RFC3339)},
		{"-10 day", true, now.Add(-10 * 24 * time.Hour).Format(time.RFC3339)},
		{"-10day", true, now.Add(-10 * 24 * time.Hour).Format(time.RFC3339)},
		{"-1 hour", true, now.Add(-1 * time.Hour).Format(time.RFC3339)},
		{"-10 hours", true, now.Add(-10 * time.Hour).Format(time.RFC3339)},
		{"-30 hour", true, now.Add(-30 * time.Hour).Format(time.RFC3339)},
		{"-30hour", true, now.Add(-30 * time.Hour).Format(time.RFC3339)},
		{"now", true, now.Format("2006-01-02T15")},
		{"", true, now.Format("2006-01-02T15")},
		{"2016/04/23", false, ""},
		{"2016/04/23 12:50", false, ""},
		{"1 day", false, ""},
		{"1 hour", false, ""},
		{"actual", false, ""},
		{strconv.FormatInt(now.Unix(), 10), true, now.Format(time.RFC3339)},
		{"wrong_timestamp", false, ""},
	}
	for index, tc := range testCases {
		tc := tc // capture range variable
		msg := fmt.Sprintf("TestCase %d", index)
		t.Run(msg, func(t *testing.T) {
			t.Parallel()

			date, err := timex.Parse(tc.dateStr)
			if tc.valid {
				assert.NoError(t, err)
				assert.Contains(t, date.Format(time.RFC3339), tc.expected, fmt.Sprintf("%s checking time", msg))
			} else {
				assert.Error(t, err, tc.dateStr, fmt.Sprintf("%s checking error", msg))
			}
		})
	}
}

func TestParseFromDate(t *testing.T) {
	testCases := []struct {
		dateStr  string
		valid    bool
		expected string
	}{
		{"2016-04-23 12:56", true, "2016-04-23T12:56:00"},
		{"2016-04-23", true, "2016-04-23T00:00:00"},
		{"2016/04/23", false, ""},
		{"2016/04/23 12:50", false, ""},
	}
	for index, tc := range testCases {
		tc := tc // capture range variable
		msg := fmt.Sprintf("TestCase %d", index)
		t.Run(msg, func(t *testing.T) {
			t.Parallel()

			date, err := timex.ParseFromDate(tc.dateStr)
			if tc.valid {
				assert.NoError(t, err)
				assert.Contains(t, date.Format(time.RFC3339), tc.expected, fmt.Sprintf("%s checking time", msg))
			} else {
				assert.Error(t, err, tc.dateStr, fmt.Sprintf("%s checking error", msg))
			}
		})
	}
}

func TestParseFromDaysAgo(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		dateStr  string
		valid    bool
		expected string
	}{
		{"-1 day", true, now.Add(-1 * 24 * time.Hour).Format(time.RFC3339)},
		{"-20 days", true, now.Add(-20 * 24 * time.Hour).Format(time.RFC3339)},
		{"-10 day", true, now.Add(-10 * 24 * time.Hour).Format(time.RFC3339)},
		{"-10day", true, now.Add(-10 * 24 * time.Hour).Format(time.RFC3339)},
		{"1 day", false, ""},
	}
	for index, tc := range testCases {
		tc := tc // capture range variable
		msg := fmt.Sprintf("TestCase %d", index)
		t.Run(msg, func(t *testing.T) {
			t.Parallel()

			date, err := timex.ParseFromDaysAgo(tc.dateStr)
			if tc.valid {
				assert.NoError(t, err)
				assert.Contains(t, date.Format(time.RFC3339), tc.expected, fmt.Sprintf("%s checking time", msg))
			} else {
				assert.Error(t, err, tc.dateStr, fmt.Sprintf("%s checking error", msg))
			}
		})
	}
}

func TestParseFromHoursAgo(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		dateStr  string
		valid    bool
		expected string
	}{
		{"-1 hour", true, now.Add(-1 * time.Hour).Format(time.RFC3339)},
		{"-10 hours", true, now.Add(-10 * time.Hour).Format(time.RFC3339)},
		{"-30 hour", true, now.Add(-30 * time.Hour).Format(time.RFC3339)},
		{"-30hour", true, now.Add(-30 * time.Hour).Format(time.RFC3339)},
		{"1 hour", false, ""},
	}
	for index, tc := range testCases {
		tc := tc // capture range variable
		msg := fmt.Sprintf("TestCase %d", index)
		t.Run(msg, func(t *testing.T) {
			t.Parallel()

			date, err := timex.ParseFromHoursAgo(tc.dateStr)
			if tc.valid {
				assert.NoError(t, err)
				assert.Contains(t, date.Format(time.RFC3339), tc.expected, fmt.Sprintf("%s checking time", msg))
			} else {
				assert.Error(t, err, tc.dateStr, fmt.Sprintf("%s checking error", msg))
			}
		})
	}
}

func TestParseFromTimestamp(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		dateStr  string
		valid    bool
		expected string
	}{
		{strconv.FormatInt(now.Unix(), 10), true, now.Format(time.RFC3339)},
		{"wrong_timestamp", false, ""},
	}
	for index, tc := range testCases {
		tc := tc // capture range variable
		msg := fmt.Sprintf("TestCase %d", index)
		t.Run(msg, func(t *testing.T) {
			t.Parallel()

			date, err := timex.ParseFromTimestamp(tc.dateStr)
			if tc.valid {
				assert.NoError(t, err)
				assert.Contains(t, date.Format(time.RFC3339), tc.expected, fmt.Sprintf("%s checking time", msg))
			} else {
				assert.Error(t, err, tc.dateStr, fmt.Sprintf("%s checking error", msg))
			}
		})
	}
}
