package datetime

import (
	"testing"
	"testing/synctest"
	"time"
)

const initialCommitTimestamp = "2023-03-22T14:48:21+01:00"

var initialCommitTime, _ = time.Parse(time.RFC3339, initialCommitTimestamp)

func TestNow(t *testing.T) {
	// Hmmm. Having written this out, it occurred to me that this test is pretty pointless but at least it'll help me
	// remember how to use synctest for something "real" in future.
	synctest.Test(t, func(t *testing.T) {
		time.Sleep(time.Until(initialCommitTime))
		synctest.Wait()

		if !Now().Equal(initialCommitTime) {
			t.Errorf("got %v, want %v", Now(), initialCommitTime)
		}
	})
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
	for b.Loop() {
		_ = Now()
	}
}

func BenchmarkTomorrow(b *testing.B) {
	for b.Loop() {
		_ = Tomorrow(initialCommitTime)
	}
}

func BenchmarkFormat(b *testing.B) {
	for b.Loop() {
		_ = Format(initialCommitTime)
	}
}

func newDateOnly(s string) time.Time {
	t, _ := time.Parse(time.DateOnly, s)
	return t
}
