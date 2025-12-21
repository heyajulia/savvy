package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"strings"

	"github.com/heyajulia/savvy/internal"
	"github.com/heyajulia/savvy/internal/config"
	"github.com/heyajulia/savvy/internal/telegram"
	"github.com/heyajulia/savvy/internal/telegram/chatid"
	"github.com/heyajulia/savvy/internal/telegram/option"
)

var (
	//go:embed templates
	templatesFS embed.FS

	templates = template.Must(template.ParseFS(templatesFS, "templates/*.tmpl"))
)

func main() {
	showVersion := flag.Bool("v", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(internal.About())
		os.Exit(0)
	}

	slog.SetDefault(internal.Logger())

	slog.Info("application info", slog.Group("app", slog.String("version", internal.Version), slog.String("commit", internal.Commit)))

	c, err := config.Read()
	if err != nil {
		slog.Error("configuration error", slog.Any("err", err))
		os.Exit(1)
	}

	token := c.Telegram.Token
	channelName := c.Telegram.ChannelName
	blueskyHandle := c.Bluesky.Handle
	lastProcessedUpdateID := int64(0)

	// TODO: I don't think it matters much in this case, but we could refactor this to use channels and goroutines.
	for {
		if err := processUpdates(token, channelName, blueskyHandle, &lastProcessedUpdateID); err != nil {
			slog.Error("could not process updates", slog.Any("err", err))
		}
	}
}

func unknownCommand(token string, userID chatid.ChatID) error {
	bot := telegram.NewClient(token)

	_, err := bot.SendMessage(userID, "Sorry, ik begrijp je niet. Probeer /start of /privacy.")

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
		option.ParseModeMarkdown,
		option.Keyboard(telegram.KeyboardPrivacy()),
	)

	return err
}

func handleCommand(token, channelName, blueskyHandle string, userID chatid.ChatID, text string) error {
	switch text {
	case "/start":
		slog.Info("received command", slog.String("command", text))

		bot := telegram.NewClient(token)

		if _, err := bot.SendMessage(
			userID,
			"Hallo! In privé-chats kan ik niet zo veel. Mijn kanaal @energieprijzen is veel interessanter.",
			option.Keyboard(telegram.KeyboardStart(channelName, blueskyHandle)),
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

func processUpdates(token, channelName, blueskyHandle string, lastProcessedUpdateID *int64) error {
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

			if err := handleCommand(token, channelName, blueskyHandle, userID, *text); err != nil {
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
