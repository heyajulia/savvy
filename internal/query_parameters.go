package internal

import (
	"net/url"
	"time"
)

func QueryParameters(t time.Time) url.Values {
	t = t.UTC()

	const rfc3339milli = "2006-01-02T15:04:05.999Z07:00"

	fromDate := time.Date(t.Year(), t.Month(), t.Day(), 22, 0, 0, 0, t.Location())
	tillDate := fromDate.AddDate(0, 0, 1).Add(-1 * time.Millisecond)

	params := url.Values{
		"fromDate":  {fromDate.Format(rfc3339milli)},
		"tillDate":  {tillDate.Format(rfc3339milli)},
		"interval":  {"4"},
		"usageType": {"1"},
		"inclBtw":   {"true"},
	}

	return params
}
