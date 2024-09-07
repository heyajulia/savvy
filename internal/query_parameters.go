package internal

import (
	"net/url"
	"time"

	"github.com/heyajulia/energieprijzen/internal/datetime"
)

func QueryParameters(t time.Time) url.Values {
	t = t.UTC()

	fromDate := time.Date(t.Year(), t.Month(), t.Day(), 22, 0, 0, 0, t.Location())
	tillDate := fromDate.AddDate(0, 0, 1).Add(-1 * time.Millisecond)

	params := url.Values{
		"fromDate":  {datetime.FormatRFC3339Milli(fromDate)},
		"tillDate":  {datetime.FormatRFC3339Milli(tillDate)},
		"interval":  {"4"},
		"usageType": {"1"},
		"inclBtw":   {"true"},
	}

	return params
}
