package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/heyajulia/energieprijzen/internal"
	"github.com/heyajulia/energieprijzen/internal/cronitor"
	"github.com/heyajulia/energieprijzen/internal/datetime"
	"github.com/heyajulia/energieprijzen/internal/mustjson"
	"github.com/heyajulia/energieprijzen/internal/ranges"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

var (
	//go:embed message.tmpl
	templateSource string

	messageTemplate = template.Must(template.New("").Parse(templateSource))
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

func readConfig() (*configuration, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}

	f, err := os.Open(filepath.Join(wd, "config.json"))
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}
	defer f.Close()

	var config configuration

	decoder := json.NewDecoder(f)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	return &config, nil
}

func main() {
	showVersion := flag.Bool("v", false, "print version and exit")
	kickstart := flag.Bool("kickstart", false, "send the energy report immediately and exit")
	flag.Parse()

	if *showVersion {
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

	log.Info("application info", slog.Group("app", slog.String("version", version), slog.String("built_at", builtAt)))

	config, err := readConfig()
	if err != nil {
		log.Error("could not read config", slog.Any("err", err))
		os.Exit(1)
	}

	token := config.Telegram.Token
	chatID := config.Telegram.ChatID.String()
	telemetryURL := config.Cronitor.TelemetryURL

	if *kickstart {
		if err := postMessage(log, token, chatID); err != nil {
			log.Error("could not post message", slog.Any("err", err))
			os.Exit(1)
		}

		os.Exit(0)
	}

	var lastPostedTime time.Time
	var lastProcessedUpdateID uint64

	// TODO: I don't think it matters much in this case, but we could refactor this to use channels and goroutines.
	for {
		if err := processUpdates(log, token, &lastProcessedUpdateID); err != nil {
			log.Error("could not process update", slog.Any("err", err))
		}

		amsterdamTime := datetime.Now()

		// The time.Since check prevents the bot from "double-posting" the energy report if the bot receives an update
		// when it's time to post the report.
		if amsterdamTime.Hour() == 15 && amsterdamTime.Minute() == 1 && time.Since(lastPostedTime) > 2*time.Minute {
			log.Info("posting energy report")

			monitor := cronitor.New(telemetryURL)

			if err := monitor.SetState(cronitor.StateRun); err != nil {
				log.Error("could not set monitor state", slog.Any("err", err), slog.Any("state", cronitor.StateRun))
			}

			state := cronitor.StateComplete

			err := postMessage(log, token, chatID)
			if err != nil {
				log.Error("could not post message", slog.Any("err", err))

				state = cronitor.StateFail
			}

			if err := monitor.SetState(state); err != nil {
				log.Error("could not set monitor state", slog.Any("err", err), slog.Any("state", state))
			}

			// I think we could use amsterdamTime here, but we use the server time here for clarity.
			lastPostedTime = time.Now()
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

func unknownCommand(log *slog.Logger, token string, userID uint64) error {
	_, err := sendTelegramRequest(log, token, "sendMessage", url.Values{
		"chat_id": {strconv.FormatUint(userID, 10)},
		"text":    {"Sorry, ik begrijp je niet. Probeer /start of /privacy."},
	})
	return err
}

func privacy(log *slog.Logger, token string, userID uint64) error {
	_, err := sendTelegramRequest(log, token, "sendMessage", url.Values{
		"chat_id": {strconv.FormatUint(userID, 10)},
		// This way we can use Markdown code formatting in a raw string literal.
		"text":         {strings.ReplaceAll(fmt.Sprintf(privacyPolicy, userID), "~", "`")},
		"parse_mode":   {"Markdown"},
		"reply_markup": {privacyReplyMarkup},
	})

	return err
}

func handleCommand(log *slog.Logger, token string, userID uint64, text string) error {
	userIDString := strconv.FormatUint(userID, 10)

	switch text {
	case "/start":
		if _, err := sendTelegramRequest(log, token, "sendMessage", url.Values{
			"chat_id":      {userIDString},
			"text":         {"Hallo! In privé-chats kan ik niet zo veel. Mijn kanaal @energieprijzen is veel interessanter."},
			"reply_markup": {startReplyMarkup},
		}); err != nil {
			return err
		}
	case "/privacy":
		if err := privacy(log, token, userID); err != nil {
			return err
		}
	default:
		if err := unknownCommand(log, token, userID); err != nil {
			return err
		}
	}

	return nil
}

func handleCallbackQuery(log *slog.Logger, token string, userID, messageID uint64, data string) error {
	userIDString := strconv.FormatUint(userID, 10)

	switch data {
	case "privacy":
		if err := privacy(log, token, userID); err != nil {
			return err
		}
	case "got_it":
		if _, err := sendTelegramRequest(log, token, "deleteMessage", url.Values{
			"chat_id":    {userIDString},
			"message_id": {strconv.FormatUint(messageID, 10)},
		}); err != nil {
			return err
		}
	default:
		if err := unknownCommand(log, token, userID); err != nil {
			return err
		}
	}

	return nil
}

func processUpdates(log *slog.Logger, token string, lastProcessedUpdateID *uint64) error {
	// TODO: Restrict allowed updates: https://core.telegram.org/bots/api#getupdates.
	resp, err := sendTelegramRequest(log, token, "getUpdates", url.Values{
		"offset":  {strconv.FormatUint(*lastProcessedUpdateID+1, 10)},
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
		*lastProcessedUpdateID = updateID

		userID, err := userID(update)
		if err != nil {
			return err
		}

		switch {
		case isMessage(update):
			text, hasText := update["message"].(map[string]any)["text"].(string)

			if !hasText {
				if err := unknownCommand(log, token, userID); err != nil {
					return err
				}

				continue
			}

			if err := handleCommand(log, token, userID, text); err != nil {
				return err
			}
		case isCallbackQuery(update):
			callbackQuery := update["callback_query"].(map[string]any)
			messageID := uint64(callbackQuery["message"].(map[string]any)["message_id"].(float64))
			data := callbackQuery["data"].(string)

			if _, err := sendTelegramRequest(log, token, "answerCallbackQuery", url.Values{
				"callback_query_id": {callbackQuery["id"].(string)},
			}); err != nil {
				return err
			}

			if err := handleCallbackQuery(log, token, userID, messageID, data); err != nil {
				return err
			}
		default:
			return errUnknownUpdateType
		}
	}

	return nil
}

func postMessage(log *slog.Logger, token, chatID string) error {
	prices, err := internal.GetEnergyPrices(log)
	if err != nil {
		return fmt.Errorf("could not get energy prices: %w", err)
	}

	var sb strings.Builder

	now := datetime.Now()
	hello, goodbye := internal.GetGreeting(now)
	tomorrow := datetime.Tomorrow(now)

	average := prices.Average()
	hourlies := make([]hourly, 24)

	for hour, price := range prices.All() {
		hourlies[hour] = hourly{
			Emoji:          internal.GetPriceEmoji(price, average),
			PaddedHour:     fmt.Sprintf("%02d", hour),
			FormattedPrice: internal.FormatCurrencyValue(price),
		}
	}

	data := templateData{
		Hello:            hello,
		Goodbye:          goodbye,
		TomorrowDate:     datetime.Format(tomorrow),
		AverageFormatted: internal.FormatCurrencyValue(average),
		AverageHours:     ranges.CollapseAndFormat(prices.AverageHours()),
		HighFormatted:    internal.FormatCurrencyValue(prices.High()),
		HighHours:        ranges.CollapseAndFormat(prices.HighHours()),
		LowFormatted:     internal.FormatCurrencyValue(prices.Low()),
		LowHours:         ranges.CollapseAndFormat(prices.LowHours()),
		Hourly:           hourlies,
	}

	if err := messageTemplate.Execute(&sb, data); err != nil {
		return fmt.Errorf("could not render report: %w", err)
	}

	message := sb.String()

	log.Info("sending message", slog.String("chat_id", chatID), slog.String("message", message))

	resp, err := sendTelegramRequest(log, token, "sendMessage", url.Values{
		"chat_id":    {chatID},
		"text":       {message},
		"parse_mode": {"HTML"},
	})
	if err != nil {
		return fmt.Errorf("could not send message: %w", err)
	}

	messageId := uint64(resp["result"].(map[string]any)["message_id"].(float64))
	idLogger := log.With(slog.Uint64("message_id", messageId))

	idLogger.Info("message sent")

	_, err = sendTelegramRequest(log, token, "setMessageReaction", url.Values{
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

	return nil
}

func sendTelegramRequest(log *slog.Logger, token, method string, params url.Values) (map[string]any, error) {
	// TODO: Maybe logging in this function is too noisy?

	// Note that we don't log the params for privacy reasons.
	log.Info("sending request to telegram", slog.String("method", method))

	resp, err := http.PostForm(fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, method), params)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	// Note that we don't log the response body for privacy reasons.
	log.Info("received response from telegram", slog.Group("response", slog.Int("status_code", resp.StatusCode)))

	var m map[string]any

	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}

	// TODO: This cast could fail. We should handle that.
	if !m["ok"].(bool) {
		description, castOk := m["description"].(string)
		if !castOk {
			description = "no description"
		}

		return nil, fmt.Errorf("telegram response not ok: %s", description)
	}

	return m, nil
}
