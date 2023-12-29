package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/heyajulia/energieprijzen/internal"
	"github.com/heyajulia/energieprijzen/internal/date"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

type credentials struct {
	Telegram   string `json:"telegram"`
	MonitorURL string `json:"cronitor_url"`
}

var (
	dryRun                    bool
	token, monitorURL, chatID string
)

const (
	stateRun      = "run"
	stateComplete = "complete"
	stateFail     = "fail"
)

func init() {
	flag.BoolVar(&dryRun, "d", false, "dry run")
	flag.Parse()

	w := os.Stderr

	log := slog.New(
		tint.NewHandler(w, &tint.Options{
			AddSource:  true,
			NoColor:    !isatty.IsTerminal(w.Fd()),
			TimeFormat: time.RFC3339,
		}),
	)

	// We do this only so we can log in this function. We pull it back out in `main`. Other functions that need to log
	// should have a logger as their (first) parameter: `log *slog.Logger`.
	slog.SetDefault(log)

	// No need for credentials or a chat ID if we're not sending a message.
	if dryRun {
		return
	}

	creds := readCredentials(log)

	token = creds.Telegram
	monitorURL = creds.MonitorURL

	if id, ok := os.LookupEnv("ENERGIEPRIJZEN_BOT_CHAT_ID"); ok {
		chatID = id
	} else {
		log.Error("ENERGIEPRIJZEN_BOT_CHAT_ID is not set")
		os.Exit(1)
	}
}

func readCredentials(log *slog.Logger) credentials {
	p := os.ExpandEnv("$CREDENTIALS_DIRECTORY/token")

	log = log.With(slog.String("path", p))

	if _, err := os.Stat(p); errors.Is(err, fs.ErrNotExist) {
		log.Error("credentials file does not exist", slog.Any("err", err))
		os.Exit(1)
	}

	b, err := os.ReadFile(p)
	if err != nil {
		log.Error("could not read credentials file", slog.Any("err", err))
		os.Exit(1)
	}

	var creds credentials

	err = json.Unmarshal(b, &creds)
	if err != nil {
		log.Error("could not decode credentials file as JSON", slog.Any("err", err))
		os.Exit(1)
	}

	return creds
}

func main() {
	log := slog.Default()

	setCronitorState(log, stateRun)

	user, err := user.Current()
	if err != nil {
		log.Error("could not get current user", slog.Any("err", err))
		setCronitorState(log, stateFail)
		os.Exit(1)
	}

	log.Info("user information", slog.Group("user", slog.String("name", user.Username), slog.String("uid", user.Uid)))

	prices, err := internal.GetEnergyPrices(log)
	if err != nil {
		log.Error("could not get energy prices", slog.Any("err", err))
		setCronitorState(log, stateFail)
		os.Exit(1)
	}

	var sb strings.Builder

	hello, goodbye := internal.GetGreeting()

	data := templateData{
		Hello:        hello,
		Goodbye:      goodbye,
		TomorrowDate: date.Tomorrow(),
		EnergyPrices: prices,
	}

	err = report(data).Render(context.Background(), &sb)
	if err != nil {
		log.Error("could not render report", slog.Any("err", err))
		setCronitorState(log, stateFail)
		os.Exit(1)
	}

	message := strings.ReplaceAll(sb.String(), "<br>", "\n")

	log.Info("sending message", slog.String("chat_id", chatID), slog.String("message", message))

	if dryRun {
		log.Info("dry run mode enabled, not sending message")
		return
	}

	resp, err := http.PostForm("https://api.telegram.org/bot"+token+"/sendMessage", url.Values{
		"chat_id":    {chatID},
		"text":       {message},
		"parse_mode": {"HTML"},
	})
	if err != nil {
		log.Error("could not send message", slog.Any("err", err))
		setCronitorState(log, stateFail)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error("unexpected status code", slog.Group("response", slog.Int("status_code", resp.StatusCode)))
		setCronitorState(log, stateFail)
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("could not read response body", slog.Any("err", err))
		setCronitorState(log, stateFail)
		os.Exit(1)
	}

	log.Info("message sent", slog.Group("response", slog.Int("status_code", resp.StatusCode), slog.String("body", string(body))))

	setCronitorState(log, stateComplete)
}

func setCronitorState(log *slog.Logger, state string) {
	if dryRun {
		return
	}

	log = log.With(slog.String("state", state))

	log.Info("setting cronitor state")

	resp, err := http.Get(fmt.Sprintf("%s?state=%s", monitorURL, state))
	if err != nil {
		log.Error("could not set cronitor state", slog.Any("err", err))
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error("unexpected status code while setting cronitor state", slog.Group("response", slog.Int("status_code", resp.StatusCode)))
		os.Exit(1)
	}
}
