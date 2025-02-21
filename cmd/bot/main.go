package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/heyajulia/energieprijzen/internal"
	"github.com/heyajulia/energieprijzen/internal/bsky"
	"github.com/heyajulia/energieprijzen/internal/config"
	"github.com/heyajulia/energieprijzen/internal/cronitor"
	"github.com/heyajulia/energieprijzen/internal/datetime"
	"github.com/heyajulia/energieprijzen/internal/prices"
	"github.com/heyajulia/energieprijzen/internal/ranges"
	"github.com/heyajulia/energieprijzen/internal/telegram"
	"github.com/heyajulia/energieprijzen/internal/telegram/chatid"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

var (
	//go:embed templates
	templatesFS embed.FS

	templates = template.Must(template.ParseFS(templatesFS, "templates/*.tmpl"))
)

func main() {
	showVersion := flag.Bool("v", false, "print version and exit")
	kickstart := flag.Bool("kickstart", false, "send the energy report immediately and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s built at %s\n", version, builtAt)
		os.Exit(0)
	}

	w := os.Stderr

	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			AddSource:  true,
			NoColor:    !isatty.IsTerminal(w.Fd()),
			TimeFormat: time.RFC3339,
		}),
	))

	slog.Info("application info", slog.Group("app", slog.String("version", version), slog.String("built_at", builtAt)))

	config, err := config.Read()
	if err != nil {
		slog.Error("configuration error", slog.Any("err", err))
		os.Exit(1)
	}

	token := config.Telegram.Token
	chatID := config.Telegram.ChatID
	blueskyIdentifier := config.Bluesky.Identifier
	blueskyPassword := config.Bluesky.Password
	cronitorURL := config.Cronitor.URL

	if *kickstart {
		if err := post(token, chatID, blueskyIdentifier, blueskyPassword); err != nil {
			slog.Error("could not post", slog.Any("err", err))
			os.Exit(1)
		}

		os.Exit(0)
	}

	var lastPostedTime time.Time
	var lastProcessedUpdateID int64

	// TODO: I don't think it matters much in this case, but we could refactor this to use channels and goroutines.
	for {
		if err := processUpdates(token, &lastProcessedUpdateID); err != nil {
			slog.Error("could not process updates", slog.Any("err", err))
		}

		amsterdamTime := datetime.Now()

		// The time.Since check prevents the bot from "double-posting" the energy report if the bot receives an update
		// when it's time to post the report.
		if amsterdamTime.Hour() == 15 && amsterdamTime.Minute() == 1 && time.Since(lastPostedTime) > 2*time.Minute {
			slog.Info("posting energy report")

			monitor := cronitor.New(cronitorURL)
			if err := monitor.Monitor(func() error {
				return post(token, chatID, blueskyIdentifier, blueskyPassword)
			}); err != nil {
				slog.Error("failed to post", slog.Any("err", err))
			}

			// I think we could use amsterdamTime here, but we use the server time here for clarity.
			lastPostedTime = time.Now()
		}
	}
}

func post(token string, chatID chatid.ChatID, blueskyIdentifier, blueskyPassword string) error {
	data, err := getTemplateData()
	if err != nil {
		return fmt.Errorf("get template data: %w", err)
	}

	url, err := postToTelegram(*data, token, chatID)
	if err != nil {
		return fmt.Errorf("post message to telegram: %w", err)
	}

	if err := postToBluesky(*data, blueskyIdentifier, blueskyPassword, url); err != nil {
		return fmt.Errorf("post to bluesky: %w", err)
	}

	return nil
}

func postToBluesky(data templateData, blueskyIdentifier, blueskyPassword, url string) error {
	client, err := bsky.Login(blueskyIdentifier, blueskyPassword)
	if err != nil {
		return fmt.Errorf("login to bluesky: %w", err)
	}

	var sb strings.Builder

	if err := templates.ExecuteTemplate(&sb, "summary.tmpl", data); err != nil {
		return fmt.Errorf("render bluesky post: %w", err)
	}

	if err := client.Post(sb.String(), url); err != nil {
		return fmt.Errorf("post to bluesky: %w", err)
	}

	return nil
}

func unknownCommand(token string, userID chatid.ChatID) error {
	bot := telegram.NewClient(token)

	_, err := bot.SendMessage(
		userID,
		"Sorry, ik begrijp je niet. Probeer /start of /privacy.",
		telegram.ParseModeDefault,
		telegram.KeyboardNone,
	)

	return err
}

func privacy(token string, userID chatid.ChatID) error {
	var sb strings.Builder

	if err := templates.ExecuteTemplate(&sb, "privacy.tmpl", userID.String()); err != nil {
		return fmt.Errorf("render privacy policy: %w", err)
	}

	bot := telegram.NewClient(token)

	_, err := bot.SendMessage(
		userID,
		sb.String(),
		telegram.ParseModeMarkdown,
		telegram.KeyboardPrivacy,
	)

	return err
}

func handleCommand(token string, userID chatid.ChatID, text string) error {
	switch text {
	case "/start":
		slog.Info("received command", slog.String("command", text))

		bot := telegram.NewClient(token)

		if _, err := bot.SendMessage(
			userID,
			"Hallo! In privé-chats kan ik niet zo veel. Mijn kanaal @energieprijzen is veel interessanter.",
			telegram.ParseModeDefault,
			telegram.KeyboardStart,
		); err != nil {
			return err
		}
	case "/privacy":
		slog.Info("received command", slog.String("command", text))

		if err := privacy(token, userID); err != nil {
			return err
		}
	default:
		slog.Info("received unknown command")

		if err := unknownCommand(token, userID); err != nil {
			return err
		}
	}

	return nil
}

func handleCallbackQuery(token string, userID chatid.ChatID, messageID int64, data string) error {
	switch data {
	case "privacy":
		slog.Info("received callback query", slog.String("data", data))

		if err := privacy(token, userID); err != nil {
			return err
		}
	case "got_it":
		slog.Info("received callback query", slog.String("data", data))

		bot := telegram.NewClient(token)

		if err := bot.DeleteMessage(userID, messageID); err != nil {
			return err
		}
	default:
		slog.Info("received unknown callback query")

		if err := unknownCommand(token, userID); err != nil {
			return err
		}
	}

	return nil
}

func processUpdates(token string, lastProcessedUpdateID *int64) error {
	bot := telegram.NewClient(token)

	updates, err := bot.GetUpdates(*lastProcessedUpdateID + 1)
	if err != nil {
		return err
	}

	for _, update := range updates {
		// This means any errors won't cause the bot to get stuck in a loop.
		*lastProcessedUpdateID = int64(update.ID)

		userID := update.UserID()

		switch {
		case update.IsMessage():
			text := update.Message.Text

			if text == nil {
				slog.Info("message doesn't contain text")

				if err := unknownCommand(token, userID); err != nil {
					return err
				}

				continue
			}

			if err := handleCommand(token, userID, *text); err != nil {
				return err
			}
		case update.IsCallbackQuery():
			callbackQuery := *update.CallbackQuery
			messageID := int64(callbackQuery.Message.ID)
			data := callbackQuery.Data

			if err := bot.AnswerCallbackQuery(callbackQuery.ID); err != nil {
				return err
			}

			if err := handleCallbackQuery(token, userID, messageID, data); err != nil {
				return err
			}
		default:
			panic("unreachable")
		}
	}

	return nil
}

func getTemplateData() (*templateData, error) {
	p, err := internal.GetEnergyPrices()
	if err != nil {
		return nil, fmt.Errorf("get energy prices: %w", err)
	}

	now := datetime.Now()
	hello, goodbye := internal.GetGreeting(now)
	tomorrow := datetime.Tomorrow(now)

	average := p.Average()
	hourlies := make([]hourly, 24)

	for hour, price := range p.All() {
		hourlies[hour] = hourly{
			Emoji:          internal.GetPriceEmoji(price, average),
			PaddedHour:     fmt.Sprintf("%02d", hour),
			FormattedPrice: prices.Format(price),
		}
	}

	data := templateData{
		Hello:            hello,
		Goodbye:          goodbye,
		TomorrowDate:     datetime.Format(tomorrow),
		AverageFormatted: prices.Format(average),
		AverageHours:     ranges.CollapseAndFormat(p.AverageHours()),
		HighFormatted:    prices.Format(p.High()),
		HighHours:        ranges.CollapseAndFormat(p.HighHours()),
		LowFormatted:     prices.Format(p.Low()),
		LowHours:         ranges.CollapseAndFormat(p.LowHours()),
		Hourly:           hourlies,
	}

	return &data, nil
}

func postToTelegram(data templateData, token string, chatID chatid.ChatID) (string, error) {
	var sb strings.Builder

	if err := templates.ExecuteTemplate(&sb, "message.tmpl", data); err != nil {
		return "", fmt.Errorf("render report: %w", err)
	}

	text := sb.String()

	slog.Info("sending message", slog.String("chat_id", chatID.String()), slog.String("message", text))

	bot := telegram.NewClient(token)

	message, err := bot.SendMessage(chatID, text, telegram.ParseModeHTML, telegram.KeyboardNone)
	if err != nil {
		return "", fmt.Errorf("send message: %w", err)
	}

	messageID := int64(message.ID)
	idLogger := slog.With(slog.Int64("message_id", messageID))

	idLogger.Info("message sent")

	if err := bot.SetMessageReaction(chatID, messageID); err != nil {
		// Not being able to react to the message is not the end of the world.
		idLogger.Warn("could not react to message", slog.Any("err", err))
	} else {
		idLogger.Info("message reacted to")
	}

	// FIXME: Harcoded channel name.
	return fmt.Sprintf("https://t.me/energieprijzen/%d", messageID), nil
}
