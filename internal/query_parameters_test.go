package internal

import (
	"testing"
	"time"
)

func TestQueryParameters(t *testing.T) {
	const initialCommitTimestamp = "2023-03-22T14:48:21+01:00"

	ts, _ := time.Parse(time.RFC3339, initialCommitTimestamp)
	q := QueryParameters(ts)

	want := map[string]string{
		"interval":  "4",
		"usageType": "1",
		"inclBtw":   "true",
		"fromDate":  "2023-03-22T22:00:00Z",
		"tillDate":  "2023-03-23T21:59:59.999Z",
	}

	for key, actual := range want {
		if got := q.Get(key); got != actual {
			t.Errorf("got %q, want %q", got, actual)
		}
	}
}
