package datetime

import (
	"testing"
	"time"
)

const initialCommitTimestamp = "2023-03-22T14:48:21+01:00"

var initialCommitTime, _ = time.Parse(time.RFC3339, initialCommitTimestamp)

func TestNow(t *testing.T) {
	oldNowFunc := now
	now = func() time.Time {
		return initialCommitTime
	}

	t.Cleanup(func() {
		now = oldNowFunc
	})

	if !Now().Equal(initialCommitTime) {
		t.Errorf("got %v, want %v", Now(), initialCommitTime)
	}
}

func TestFormat(t *testing.T) {
	tm := newDateOnly("2009-11-10")

	got := Format(tm)
	want := "dinsdag 10 november 2009"

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTomorrow(t *testing.T) {
	tm := newDateOnly("2000-01-01")

	got := Tomorrow(tm)
	want := newDateOnly("2000-01-02")

	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func BenchmarkNow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Now()
	}
}

func BenchmarkTomorrow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Tomorrow(initialCommitTime)
	}
}

func BenchmarkFormat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Format(initialCommitTime)
	}
}

func newDateOnly(s string) time.Time {
	t, _ := time.Parse(time.DateOnly, s)
	return t
}
