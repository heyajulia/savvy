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
	"github.com/heyajulia/energieprijzen/internal/mustjson"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

type credentials struct {
	Telegram   string `json:"telegram"`
	MonitorURL string `json:"cronitor,omitempty"`
}

var (
	dryRun, showVersion       bool
	token, monitorURL, chatID string
	lastProcessedUpdateID     uint64
)

var postMessageReaction = mustjson.Encode([]map[string]string{{"type": "emoji", "emoji": "⚡"}})

var (
	startReplyMarkup = mustjson.Encode(map[string]any{
		"inline_keyboard": [][]map[string]string{
			{{"text": "Lees hoe ik met je privacy omga", "callback_data": "privacy"}},
			{{"text": "Abonneer je op mijn kanaal", "url": "https://t.me/energieprijzen"}},
		},
	})
	privacyReplyMarkup = mustjson.Encode(map[string]any{
		"inline_keyboard": [][]map[string]string{
			{{"text": "Verwijder dit bericht", "callback_data": "got_it"}},
		},
	})
)

var errUnknownUpdateType = errors.New("unknown update type")

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
		log.Warn("dry run mode is enabled; will not send messages or ping Cronitor")
	}

	u, err := user.Current()
	if err != nil {
		log.Error("could not get current user", slog.Any("err", err))
		os.Exit(1)
	}

	log.Info("user information", slog.Group("user", slog.String("name", u.Username), slog.String("uid", u.Uid)))

	// TODO: I don't think it matters much in this case, but we could refactor this to use channels and goroutines.
	for {
		if err := processUpdates(log); err != nil {
			log.Error("could not process update", slog.Any("err", err))
		}

		amsterdamTime := date.Amsterdam()

		// I found the docs quite confusing, but this is correct; see https://go.dev/play/p/iJ98dWNp6R7.
		// TODO: Make this easier to test in dry run mode.
		if amsterdamTime.Hour() == 15 && amsterdamTime.Minute() == 1 {
			postMessage(log)
		}
	}
}

func isMessage(update map[string]any) bool {
	_, hasMessage := update["message"]
	return hasMessage
}

func isCallbackQuery(update map[string]any) bool {
	_, hasCallbackQuery := update["callback_query"]
	return hasCallbackQuery
}

func userID(update map[string]any) (uint64, error) {
	var typ string

	switch {
	case isMessage(update):
		typ = "message"
	case isCallbackQuery(update):
		typ = "callback_query"
	default:
		return 0, errUnknownUpdateType
	}

	return uint64(update[typ].(map[string]any)["from"].(map[string]any)["id"].(float64)), nil
}

func unknownCommand(log *slog.Logger, userID uint64) error {
	_, err := doTelegramRequest(log, "sendMessage", url.Values{
		"chat_id": {strconv.FormatUint(userID, 10)},
		"text":    {"Sorry, ik begrijp je niet. Probeer /start of /privacy."},
	})
	return err
}

func privacy(log *slog.Logger, userID uint64) error {
	_, err := doTelegramRequest(log, "sendMessage", url.Values{
		"chat_id": {strconv.FormatUint(userID, 10)},
		// This way we can use Markdown code formatting in a raw string literal.
		"text":         {strings.ReplaceAll(fmt.Sprintf(privacyPolicy, userID), "~", "`")},
		"parse_mode":   {"Markdown"},
		"reply_markup": {privacyReplyMarkup},
	})

	return err
}

func handleCommand(log *slog.Logger, userID uint64, text string) error {
	userIDString := strconv.FormatUint(userID, 10)

	switch text {
	case "/start":
		if _, err := doTelegramRequest(log, "sendMessage", url.Values{
			"chat_id":      {userIDString},
			"text":         {"Hallo! In privé-chats kan ik niet zo veel. Mijn kanaal @energieprijzen is veel interessanter."},
			"reply_markup": {startReplyMarkup},
		}); err != nil {
			return err
		}
	case "/privacy":
		if err := privacy(log, userID); err != nil {
			return err
		}
	default:
		if err := unknownCommand(log, userID); err != nil {
			return err
		}
	}

	return nil
}

func handleCallbackQuery(log *slog.Logger, userID, messageID uint64, data string) error {
	userIDString := strconv.FormatUint(userID, 10)

	switch data {
	case "privacy":
		if err := privacy(log, userID); err != nil {
			return err
		}
	case "got_it":
		if _, err := doTelegramRequest(log, "deleteMessage", url.Values{
			"chat_id":    {userIDString},
			"message_id": {strconv.FormatUint(messageID, 10)},
		}); err != nil {
			return err
		}
	default:
		if err := unknownCommand(log, userID); err != nil {
			return err
		}
	}

	return nil
}

func processUpdates(log *slog.Logger) error {
	// TODO: Restrict allowed updates: https://core.telegram.org/bots/api#getupdates.
	resp, err := doTelegramRequest(log, "getUpdates", url.Values{
		"offset":  {strconv.FormatUint(lastProcessedUpdateID+1, 10)},
		"timeout": {"60"},
	})
	if err != nil {
		return err
	}

	updates := resp["result"].([]any)

	for _, update := range updates {
		update := update.(map[string]any)
		updateID := uint64(update["update_id"].(float64))

		// This means any errors won't cause the bot to get stuck in a loop.
		lastProcessedUpdateID = updateID

		userID, err := userID(update)
		if err != nil {
			return err
		}

		switch {
		case isMessage(update):
			text, hasText := update["message"].(map[string]any)["text"].(string)

			if !hasText {
				if err := unknownCommand(log, userID); err != nil {
					return err
				}

				continue
			}

			if err := handleCommand(log, userID, text); err != nil {
				return err
			}
		case isCallbackQuery(update):
			callbackQuery := update["callback_query"].(map[string]any)
			messageID := uint64(callbackQuery["message"].(map[string]any)["message_id"].(float64))
			data := callbackQuery["data"].(string)

			if _, err := doTelegramRequest(log, "answerCallbackQuery", url.Values{
				"callback_query_id": {callbackQuery["id"].(string)},
			}); err != nil {
				return err
			}

			if err := handleCallbackQuery(log, userID, messageID, data); err != nil {
				return err
			}
		default:
			return errUnknownUpdateType
		}
	}

	return nil
}

func postMessage(log *slog.Logger) {
	pingCronitor(log, stateRun)

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

	resp, err := doTelegramRequest(log, "sendMessage", url.Values{
		"chat_id":    {chatID},
		"text":       {message},
		"parse_mode": {"HTML"},
	})
	if err != nil {
		log.Error("could not send message", slog.Any("err", err))
		pingCronitor(log, stateFail)
		os.Exit(1)
	}

	pingCronitor(log, stateComplete)

	messageId := uint64(resp["result"].(map[string]any)["message_id"].(float64))
	idLogger := log.With(slog.Uint64("message_id", messageId))

	idLogger.Info("message sent")

	_, err = doTelegramRequest(log, "setMessageReaction", url.Values{
		"chat_id":    {chatID},
		"message_id": {strconv.FormatUint(messageId, 10)},
		"is_big":     {"true"},
		"reaction":   {postMessageReaction},
	})
	if err != nil {
		// Not being able to react to the message is not a fatal error because it's not an essential feature.
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
