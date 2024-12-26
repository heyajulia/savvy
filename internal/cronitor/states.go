package cronitor

var (
	stateRun      = state{"run"}
	stateComplete = state{"complete"}
	stateFail     = state{"fail"}
)

type state struct {
	state string
}

func (s state) String() string {
	return s.state
}
