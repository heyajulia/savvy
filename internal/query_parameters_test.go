package internal

import (
	"testing"
	"time"
)

func TestQueryParameters(t *testing.T) {
	const initialCommitTimestamp = "2023-03-22T14:48:21+01:00"

	ts, _ := time.Parse(time.RFC3339, initialCommitTimestamp)
	q := QueryParameters(ts)

	if got, want := q.Get("interval"), "4"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	if got, want := q.Get("usageType"), "1"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	if got, want := q.Get("inclBtw"), "true"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	if got, want := q.Get("fromDate"), "2023-03-22T22:00:00Z"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	if got, want := q.Get("tillDate"), "2023-03-23T21:59:59.999Z"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
