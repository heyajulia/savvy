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
		{nil, StateRun, true},
		{nil, StateComplete, false},
		{nil, StateFail, false},

		{&StateRun, StateRun, false},
		{&StateRun, StateComplete, true},
		{&StateRun, StateFail, true},

		{&StateComplete, StateRun, false},
		{&StateComplete, StateComplete, false},
		{&StateComplete, StateFail, false},

		{&StateFail, StateRun, false},
		{&StateFail, StateComplete, false},
		{&StateFail, StateFail, false},
	}

	for _, tt := range states {
		if got := stateTransitionAllowed(tt.from, tt.to); got != tt.want {
			t.Errorf("stateTransitionAllowed(%s, %s) = %v, want %v", tt.from, tt.to, got, tt.want)
		}
	}
}
