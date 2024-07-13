package cronitor

var (
	StateRun      = state{"run"}
	StateComplete = state{"complete"}
	StateFail     = state{"fail"}
)

type state struct {
	state string
}

func (s state) String() string {
	return s.state
}
