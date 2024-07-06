package main

import (
	"bytes"
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
	Telegram string `json:"telegram"`
}

var (
	dryRun, showVersion   bool
	token, chatID         string
	lastProcessedUpdateID uint64
)

func init() {
	flag.BoolVar(&dryRun, "d", false, "dry run")
	flag.StringVar(&token, "t", "", "Telegram bot token")
	flag.BoolVar(&showVersion, "v", false, "print version and exit")
	flag.Parse()

	if showVersion {
		fmt.Printf("%s built at %s\n", version, builtAt)
		os.Exit(0)
	}

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

		return
	}

	creds := readCredentials(log)

	token = creds.Telegram
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

	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.DisallowUnknownFields()

	err = decoder.Decode(&creds)
	if err != nil {
		log.Error("could not decode credentials file as JSON", slog.Any("err", err))
		os.Exit(1)
	}

	return creds
}

func main() {
	log := slog.Default()

	if dryRun {
		log.Warn("dry run mode is enabled; will not send messages")
	}

	u, err := user.Current()
	if err != nil {
		log.Error("could not get current user", slog.Any("err", err))
		os.Exit(1)
	}

	log.Info("user information", slog.Group("user", slog.String("name", u.Username), slog.String("uid", u.Uid)))

	// TODO: I don't think it matters much in this case, but we could refactor this to use channels and goroutines.
	for {
		processUpdates(log)

		amsterdamTime := date.Amsterdam()

		// I found the docs quite confusing, but this is correct; see https://go.dev/play/p/iJ98dWNp6R7.
		// TODO: Make this easier to test in dry run mode.
		if amsterdamTime.Hour() == 15 && amsterdamTime.Minute() == 1 {
			postMessage(log)
		}

		time.Sleep(2 * time.Second)
	}

}

func processUpdates(log *slog.Logger) error {
	resp, err := doTelegramRequest(log, "getUpdates", url.Values{
		"offset": {strconv.FormatUint(lastProcessedUpdateID+1, 10)},
	})
	if err != nil {
		return err
	}

	updates := resp["result"].([]any)

	for _, update := range updates {
		updateID := uint64(update.(map[string]any)["update_id"].(float64))
		userID := uint64(update.(map[string]any)["message"].(map[string]any)["from"].(map[string]any)["id"].(float64))
		userIDString := strconv.FormatUint(userID, 10)
		text, hasText := update.(map[string]any)["message"].(map[string]any)["text"].(string)

		// This means any errors won't cause the bot to get stuck in a loop.
		lastProcessedUpdateID = updateID

		if !hasText {
			_, err := doTelegramRequest(log, "sendMessage", url.Values{
				"chat_id": {userIDString},
				"text":    {"Sorry, ik begrijp je niet. Probeer /start of /privacy."},
			})
			if err != nil {
				return err
			}

			continue
		}

		switch text {
		case "/start":
			_, err := doTelegramRequest(log, "sendMessage", url.Values{
				"chat_id": {userIDString},
				"text":    {"Hallo! In privé-chats kan ik niet zo veel. Mijn kanaal @energieprijzen is veel interessanter. Je kunt via /privacy lezen hoe ik met je gegevens omga."},
				// TODO: Add inline keyboard with "Privacy" and "Subscribe" buttons?
			})
			if err != nil {
				return err
			}
		case "/privacy":
			_, err := doTelegramRequest(log, "sendMessage", url.Values{
				"chat_id": {userIDString},
				// This way we can use Markdown code formatting in a raw string literal.
				"text":       {strings.ReplaceAll(fmt.Sprintf(privacyPolicy, userID), "~", "`")},
				"parse_mode": {"Markdown"},
			})
			if err != nil {
				return err
			}
		default:
			_, err := doTelegramRequest(log, "sendMessage", url.Values{
				"chat_id": {userIDString},
				"text":    {"Sorry, ik begrijp je niet. Probeer /start of /privacy."},
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func postMessage(log *slog.Logger) {
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

	log.Info("sending message", slog.String("chat_id", chatID), slog.String("message", message))

	resp, err := doTelegramRequest(log, "sendMessage", url.Values{
		"chat_id":    {chatID},
		"text":       {message},
		"parse_mode": {"HTML"},
	})
	if err != nil {
		log.Error("could not send message", slog.Any("err", err))
		os.Exit(1)
	}

	messageId := uint64(resp["result"].(map[string]any)["message_id"].(float64))
	idLogger := log.With(slog.Uint64("message_id", messageId))

	idLogger.Info("message sent")

	_, err = doTelegramRequest(log, "setMessageReaction", url.Values{
		"chat_id":    {chatID},
		"message_id": {strconv.FormatUint(messageId, 10)},
		"is_big":     {"true"},
		"reaction":   {`[{"type":"emoji","emoji":"⚡"}]`},
	})
	if err != nil {
		// Not being able to react to the message is not a fatal error. Users have already received the message, so our
		// job is done.
		idLogger.Warn("could not react to message", slog.Any("err", err))
	} else {
		idLogger.Info("message reacted to")
	}

}

func doTelegramRequest(log *slog.Logger, method string, params url.Values) (map[string]any, error) {
	// TODO: Maybe logging in this function is too noisy?

	// Note that we don't log the params for privacy reasons.
	log.Info("sending telegram request", slog.String("method", method))

	if dryRun {
		switch method {
		case "sendMessage":
			// This is just enough of a response to make the bot think it sent a message.
			return map[string]any{
				"result": map[string]any{
					"message_id": float64(0),
				},
			}, nil
		case "getUpdates":
			return map[string]any{
				"result": []any{},
			}, nil
		case "setMessageReaction":
			// We don't have to mock a response because we never check it.
			return nil, nil
		default:
			panic(fmt.Sprintf("unknown method: %s", method))
		}
	}

	resp, err := http.PostForm(fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, method), params)
	if err != nil {
		return nil, fmt.Errorf("could not send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}

	// Note that we don't log the response body for privacy reasons.
	log.Info("received response from telegram", slog.Group("response", slog.Int("status_code", resp.StatusCode)))

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
