package internal

import "fmt"

var (
	Version = "v0.0.0"
	Commit  = "0000000000000000000000000000000000000000"
)

func About() string {
	return fmt.Sprintf("%s built from %s", Version, Commit)
}
