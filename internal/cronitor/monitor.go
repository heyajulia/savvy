package cronitor

import (
	"errors"
	"fmt"
	"net/http"
	neturl "net/url"
)

var ErrUnexpectedStatusCode = errors.New("cronitor: received unexpected HTTP status code")

// Monitor represents a Cronitor job monitor.
//
// A Monitor with an empty URL behaves as a "no-op Monitor" and does not make HTTP requests.
type Monitor struct {
	url   string
	state *state
}

// New creates a new Monitor with the given Cronitor telemetry URL. Panics if "net/url".Parse fails.
//
// Passing the empty string will create a "no-op Monitor" that does not make any HTTP requests.
func New(url string) *Monitor {
	_, err := neturl.Parse(url)
	if err != nil {
		// Passing the empty string does not cause an error, so we shouldn't hit this when creating a "no-op Monitor".
		panic(fmt.Sprintf("cronitor: invalid URL: %s", err))
	}

	return &Monitor{url: url}
}

// SetState sets the state of the monitor. Panics if the state transition is not allowed.
//
// If the HTTP request fails or the response status code is not OK, SetState returns an error and the internal state
// will be reverted to the previous state. If the Monitor is a "no-op Monitor" (i.e. the URL is the empty string),
// SetState will still advance the internal state of the Monitor.
func (c *Monitor) SetState(state state) error {
	if !stateTransitionAllowed(c.state, state) {
		panic(fmt.Sprintf("cronitor: cannot set a job from '%s' to '%s'", c.state, state))
	}

	prevState := c.state
	c.state = &state

	if c.url == "" {
		return nil
	}

	resp, err := http.Get(fmt.Sprintf("%s?state=%s", c.url, state))
	if err != nil {
		c.state = prevState
		return fmt.Errorf("cronitor: failed to set state: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.state = prevState
		return ErrUnexpectedStatusCode
	}

	return nil
}

func stateTransitionAllowed(from *state, to state) bool {
	if from == nil {
		return to == StateRun
	}

	if *from == StateRun {
		return to == StateComplete || to == StateFail
	}

	return false
}
