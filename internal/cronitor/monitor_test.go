package cronitor

import (
	"testing"
)

func TestStateTransitionAllowed(t *testing.T) {
	states := []struct {
		from *state
		to   state
		want bool
	}{
		{nil, stateRun, true},
		{nil, stateComplete, false},
		{nil, stateFail, false},

		{&stateRun, stateRun, false},
		{&stateRun, stateComplete, true},
		{&stateRun, stateFail, true},

		{&stateComplete, stateRun, false},
		{&stateComplete, stateComplete, false},
		{&stateComplete, stateFail, false},

		{&stateFail, stateRun, false},
		{&stateFail, stateComplete, false},
		{&stateFail, stateFail, false},
	}

	for _, tt := range states {
		if got := stateTransitionAllowed(tt.from, tt.to); got != tt.want {
			t.Errorf("stateTransitionAllowed(%s, %s) = %v, want %v", tt.from, tt.to, got, tt.want)
		}
	}
}
