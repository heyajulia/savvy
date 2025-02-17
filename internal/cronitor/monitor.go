package cronitor

import (
	"errors"
	"fmt"
	"net/http"
	neturl "net/url"
)

var (
	stateRun      = "run"
	stateComplete = "complete"
	stateFail     = "fail"

	errUnexpectedStatusCode = errors.New("received unexpected HTTP status code")
)

// Monitor represents a Cronitor job monitor.
//
// A Monitor with an empty URL behaves as a "no-op Monitor" and does not make HTTP requests.
type Monitor struct {
	url   string
	state *string
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

// Monitor invokes f and communicates its state to Cronitor.
//
// Monitor propagates errors related to setting the state, but in case f fails and setting the state fails, the state
// error is lost.
func (c *Monitor) Monitor(f func() error) error {
	if err := c.setState(stateRun); err != nil {
		return fmt.Errorf("cronitor: set state to 'run': %w", err)
	}

	if err := f(); err != nil {
		if stateErr := c.setState(stateFail); stateErr != nil {
			return fmt.Errorf("cronitor: monitored function: %w; set state to 'fail': %v", err, stateErr)
		}

		return fmt.Errorf("cronitor: monitored function: %w", err)
	}

	if err := c.setState(stateComplete); err != nil {
		return fmt.Errorf("cronitor: set state to 'complete': %w", err)
	}

	return nil
}

// SetState sets the state of the monitor. Panics if the state transition is not allowed.
//
// If the HTTP request fails or the response status code is not OK, SetState returns an error and the internal state
// will be reverted to the previous state. If the Monitor is a "no-op Monitor" (i.e. the URL is the empty string),
// SetState will still advance the internal state of the Monitor.
func (c *Monitor) setState(state string) error {
	if !stateTransitionAllowed(c.state, state) {
		panic(fmt.Sprintf("cannot set a job from '%v' to '%s'", c.state, state))
	}

	prevState := c.state
	c.state = &state

	if c.url == "" {
		return nil
	}

	resp, err := http.Get(fmt.Sprintf("%s?state=%s", c.url, state))
	if err != nil {
		c.state = prevState
		return fmt.Errorf("failed to set state: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.state = prevState
		return errUnexpectedStatusCode
	}

	return nil
}

func stateTransitionAllowed(from *string, to string) bool {
	if from == nil {
		return to == stateRun
	}

	if *from == stateRun {
		return to == stateComplete || to == stateFail
	}

	return false
}
