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
	"strconv"
	"strings"
	"time"

	"github.com/heyajulia/energieprijzen/internal"
	"github.com/heyajulia/energieprijzen/internal/date"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

type credentials struct {
	Telegram   string `json:"telegram"`
	MonitorURL string `json:"cronitor_url,omitempty"`
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
	flag.StringVar(&token, "t", "", "Telegram bot token")
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

	if id, ok := os.LookupEnv("ENERGIEPRIJZEN_BOT_CHAT_ID"); ok {
		chatID = id
	} else {
		log.Error("ENERGIEPRIJZEN_BOT_CHAT_ID is not set")
		os.Exit(1)
	}

	if token != "" {
		log.Warn("using token from flag. do not use this flag in production.")

		// As a consequence of exiting here, Cronitor will not be notified of the state of the job. That's fine, though;
		// this flag is only for local testing.

		return
	}

	creds := readCredentials(log)

	token = creds.Telegram
	monitorURL = creds.MonitorURL
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

	pingCronitor(log, stateRun)

	u, err := user.Current()
	if err != nil {
		log.Error("could not get current user", slog.Any("err", err))
		pingCronitor(log, stateFail)
		os.Exit(1)
	}

	log.Info("user information", slog.Group("user", slog.String("name", u.Username), slog.String("uid", u.Uid)))

	prices, err := internal.GetEnergyPrices(log)
	if err != nil {
		log.Error("could not get energy prices", slog.Any("err", err))
		pingCronitor(log, stateFail)
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
		pingCronitor(log, stateFail)
		os.Exit(1)
	}

	message := strings.ReplaceAll(sb.String(), "<br>", "\n")

	log.Info("sending message", slog.String("chat_id", chatID), slog.String("message", message))

	if dryRun {
		log.Info("dry run mode enabled, not sending message")
		return
	}

	resp, err := doTelegramRequest("sendMessage", url.Values{
		"chat_id":    {chatID},
		"text":       {message},
		"parse_mode": {"HTML"},
	})
	if err != nil {
		log.Error("could not send message", slog.Any("err", err))
		pingCronitor(log, stateFail)
		os.Exit(1)
	}

	messageId := uint64(resp["result"].(map[string]any)["message_id"].(float64))

	log.Info("message sent", slog.Uint64("message_id", messageId))

	_, err = doTelegramRequest("setMessageReaction", url.Values{
		"chat_id":    {chatID},
		"message_id": {strconv.FormatUint(messageId, 10)},
		"is_big":     {"true"},
		"reaction":   {`[{"type":"emoji","emoji":"⚡"}]`},
	})
	if err != nil {
		// Not being able to react to the message is not a fatal error. The users have already received the message, so
		// our job is done.
		log.Warn("could not react to message", slog.Uint64("message_id", messageId), slog.Any("err", err))
	} else {
		log.Info("message reacted to", slog.Uint64("message_id", messageId))
	}

	pingCronitor(log, stateComplete)
}

func pingCronitor(log *slog.Logger, state string) {
	log = log.With(slog.String("state", state))

	if dryRun || monitorURL == "" {
		return
	}

	log.Info("pinging cronitor")

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

func doTelegramRequest(method string, params url.Values) (map[string]any, error) {
	resp, err := http.PostForm(fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, method), params)
	if err != nil {
		return nil, fmt.Errorf("could not send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var m map[string]any

	err = json.Unmarshal(body, &m)
	if err != nil {
		return nil, fmt.Errorf("could not decode response body as JSON: %w", err)
	}

	if !m["ok"].(bool) {
		description, castOk := m["description"].(string)
		if !castOk {
			description = "no description"
		}

		return nil, fmt.Errorf("telegram error: %s", description)
	}

	return m, nil
}
