package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/heyajulia/energieprijzen/internal"
	"github.com/heyajulia/energieprijzen/internal/date"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

type credentials struct {
	Telegram string `json:"telegram"`
}

var (
	dryRun        bool
	token, chatID string
)

func init() {
	flag.BoolVar(&dryRun, "d", false, "dry run")
	flag.Parse()

	w := os.Stderr

	log := slog.New(
		tint.NewHandler(w, &tint.Options{
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

	f, err := os.Open(p)
	if err != nil {
		log.Error("could not open credentials file", slog.Any("err", err))
		os.Exit(1)
	}
	defer f.Close()

	var creds credentials

	err = json.NewDecoder(f).Decode(&creds)
	if err != nil {
		log.Error("could not decode credentials file as JSON", slog.Any("err", err))
		os.Exit(1)
	}

	return creds
}

func main() {
	log := slog.Default()

	user, err := user.Current()
	if err != nil {
		log.Error("could not get current user", slog.Any("err", err))
		os.Exit(1)
	}

	log.Info("program starting", slog.Group("user", slog.String("name", user.Username), slog.String("uid", user.Uid)))

	if dryRun {
		log.Info("dry run mode enabled")
	}

	prices, err := internal.GetEnergyPrices(log)
	if err != nil {
		log.Error("could not get energy prices", slog.Any("err", err))
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
		os.Exit(1)
	}

	message := strings.ReplaceAll(sb.String(), "<br>", "\n")

	if dryRun {
		return
	}

	log.Info("sending message", slog.String("chat_id", chatID), slog.String("message", message))

	resp, err := http.PostForm("https://api.telegram.org/bot"+token+"/sendMessage", url.Values{
		"chat_id":    {chatID},
		"text":       {message},
		"parse_mode": {"HTML"},
	})
	if err != nil {
		log.Error("could not send message", slog.Any("err", err))
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error("unexpected status code", slog.Int("status_code", resp.StatusCode))
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("could not read response body", slog.Any("err", err))
		os.Exit(1)
	}

	log.Debug("message sent", slog.Group("response", slog.Int("status_code", resp.StatusCode), slog.String("body", string(body))))
}
