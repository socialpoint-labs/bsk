package timex

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	regexDateYYYYMMDDhm = regexp.MustCompile(`^(20[0-9]{2})-([0-9]{2})-([0-9]{2})(\s([0-9]{1,2}):([0-9]{2}))?`)
	regexDateDays       = regexp.MustCompile(`^-([0-9]{1,2})\s?days?`)
	regexDateHours      = regexp.MustCompile(`^-([0-9]{1,2})\s?hours?`)
	regexTimestamp      = regexp.MustCompile(`^([0-9]{10})`)
)

// Parse returns a new Time from string
//
// Example of valid inputs:
// - Current time: "now" or ""
// - Dates: "2016-04-23 12:56" or "2016-04-23"
// - Days: "-1 day" or "-10 days"
// - Hours: "-1 hour" or "-10 hours"
// - Timestamp: "1464876005"
func Parse(dateStr string) (time.Time, error) {
	if dateStr == "" || strings.ToLower(dateStr) == "now" {
		return time.Now(), nil
	}

	match := regexDateYYYYMMDDhm.MatchString(dateStr)
	if match {
		return ParseFromDate(dateStr)
	}

	match = regexDateDays.MatchString(dateStr)
	if match {
		return ParseFromDaysAgo(dateStr)
	}

	match = regexDateHours.MatchString(dateStr)
	if match {
		return ParseFromHoursAgo(dateStr)
	}

	match = regexTimestamp.MatchString(dateStr)
	if match {
		return ParseFromTimestamp(dateStr)
	}

	return time.Time{}, fmt.Errorf("Invalid date format: %s", dateStr)
}

// ParseFromDate returns Time from string with days
//
// Example of valid inputs: "2016-04-23 12:56" or "2016-04-23"
func ParseFromDate(dateStr string) (time.Time, error) {
	res := regexDateYYYYMMDDhm.FindStringSubmatch(dateStr)
	if len(res) < 4 {
		return time.Time{}, fmt.Errorf("Invalid date format: %s", dateStr)
	}

	y, err := strconv.Atoi(res[1])
	if err != nil {
		return time.Time{}, err
	}
	m, err := strconv.Atoi(res[2])
	if err != nil {
		return time.Time{}, err
	}
	d, err := strconv.Atoi(res[3])
	if err != nil {
		return time.Time{}, err
	}
	h, err := strconv.Atoi(res[5])
	if err != nil {
		h = 0
	}
	mn, err := strconv.Atoi(res[6])
	if err != nil {
		mn = 0
	}

	return time.Date(y, time.Month(m), d, h, mn, 0, 0, time.UTC), nil
}

// ParseFromDaysAgo returns Time from string with days
//
// Example of valid inputs: "-1 day" or "-10 days"
func ParseFromDaysAgo(dateStr string) (time.Time, error) {
	res := regexDateDays.FindStringSubmatch(dateStr)
	if len(res) != 2 {
		return time.Time{}, fmt.Errorf("Invalid date format: %s", dateStr)
	}

	nbDay, err := strconv.Atoi(res[1])
	if err != nil {
		return time.Time{}, err
	}

	nbSecond := int64(nbDay * 24 * 60 * 60)
	timestamp := time.Now().Unix() - nbSecond

	return time.Unix(timestamp, 0), nil
}

// ParseFromHoursAgo returns Time from string with hours
//
// Example of valid inputs: "-1 hour" or "-10 hours"
func ParseFromHoursAgo(dateStr string) (time.Time, error) {
	res := regexDateHours.FindStringSubmatch(dateStr)
	if len(res) != 2 {
		return time.Time{}, fmt.Errorf("Invalid date format: %s", dateStr)
	}

	nbHours, err := strconv.Atoi(res[1])
	if err != nil {
		return time.Time{}, err
	}

	nbSecond := int64(nbHours * 60 * 60)
	timestamp := time.Now().Unix() - nbSecond

	return time.Unix(timestamp, 0), nil
}

// ParseFromTimestamp returns Time from a timestamp string
//
// Example of valid inputs: "1464876005"
func ParseFromTimestamp(timestamp string) (time.Time, error) {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(ts, 0), nil
}
