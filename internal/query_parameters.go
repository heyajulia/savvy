package internal

import (
	"fmt"
	"net/url"
	"time"

	"github.com/heyajulia/energieprijzen/internal/datetime"
)

// QueryParameters returns the query parameters to retrieve tomorrow's prices.
func QueryParameters(t time.Time) (url.Values, error) {
	loc, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		return nil, fmt.Errorf("load time zone info: %w", err)
	}

	tAmsterdam := t.In(loc)
	tomorrow := tAmsterdam.AddDate(0, 0, 1)

	fromDateLocal := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, loc)
	tillDateLocal := fromDateLocal.AddDate(0, 0, 1).Add(-time.Millisecond)

	// Convert the local boundaries to UTC.
	params := url.Values{
		"fromDate":  {datetime.FormatRFC3339Milli(fromDateLocal.UTC())},
		"tillDate":  {datetime.FormatRFC3339Milli(tillDateLocal.UTC())},
		"interval":  {"4"},
		"usageType": {"1"},
		"inclBtw":   {"true"},
	}

	return params, nil
}
