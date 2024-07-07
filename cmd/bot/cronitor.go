package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

type state int

const (
	stateRun state = iota
	stateComplete
	stateFail
)

func pingCronitor(log *slog.Logger, s state) {
	var state string

	switch s {
	case stateRun:
		state = "run"
	case stateComplete:
		state = "complete"
	case stateFail:
		state = "fail"
	default:
		// Calling the function with an unknown state is a programming error, so let's fail loudly.
		panic(fmt.Sprintf("unknown state: %d", s))
	}

	log = log.With(slog.String("state", state))

	log.Info("pinging cronitor")

	if dryRun {
		return
	}

	if monitorURL == "" {
		log.Warn("no cronitor URL set, not pinging cronitor")
		return
	}

	resp, err := http.Get(fmt.Sprintf("%s?state=%s", monitorURL, state))
	if err != nil {
		log.Error("could not ping cronitor", slog.Any("err", err))
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error("unexpected status code while pinging cronitor", slog.Group("response", slog.Int("status_code", resp.StatusCode)))
		os.Exit(1)
	}
}
