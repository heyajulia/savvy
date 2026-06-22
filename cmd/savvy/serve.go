package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/heyajulia/savvy/internal"
	"github.com/heyajulia/savvy/internal/config"
	"github.com/heyajulia/savvy/internal/telegram"
	"github.com/heyajulia/savvy/internal/telegram/option"
	"github.com/urfave/cli/v3"
)

func serveCommand() *cli.Command {
	return &cli.Command{
		Name:   "serve",
		Usage:  "Start the Savvy Telegram bot server",
		Action: runServe,
	}
}

func runServe(ctx context.Context, c *cli.Command) error {
	slog.SetDefault(internal.Logger())

	slog.Info("application info", slog.Group("app", slog.String("version", internal.Version), slog.String("commit", internal.Commit)))

	cfg, err := config.Read[config.Serve]()
	if err != nil {
		slog.Error("configuration error", slog.Any("err", err))
		os.Exit(1)
	}

	token := cfg.Telegram.Token
	channelName := cfg.Telegram.ChannelName
	blueskyIdentifier := cfg.Bluesky.Identifier

	bot := telegram.NewBot(token)

	bot.OnCommand("/start", func(ctx *telegram.Context) error {
		slog.Info("received command", slog.String("command", "/start"))
		return ctx.Reply(
			"Hallo! In privé-chats kan ik niet zo veel. Mijn kanaal @energieprijzen is veel interessanter.",
			option.Keyboard(telegram.KeyboardStart(channelName, blueskyIdentifier)),
		)
	})

	bot.OnCommand("/privacy", func(ctx *telegram.Context) error {
		slog.Info("received command", slog.String("command", "/privacy"))
		return handlePrivacy(ctx)
	})

	bot.OnCallback("privacy", func(ctx *telegram.Context) error {
		slog.Info("received callback query", slog.String("data", "privacy"))
		return handlePrivacy(ctx)
	})

	bot.OnCallback("got_it", func(ctx *telegram.Context) error {
		slog.Info("received callback query", slog.String("data", "got_it"))
		return ctx.DeleteMessage(ctx.Update().CallbackQuery.Message.ID)
	})

	bot.OnUnknown(func(ctx *telegram.Context) error {
		return ctx.Reply("Sorry, ik begrijp je niet. Probeer /start of /privacy.")
	})

	lastProcessedUpdateID := int64(0)

	for {
		select {
		case <-ctx.Done():
			slog.Info("shutting down")
			return nil
		default:
			id, err := bot.ProcessUpdates(lastProcessedUpdateID + 1)
			if err != nil {
				slog.Error("could not process updates", slog.Any("err", err))
			}
			if id > lastProcessedUpdateID {
				lastProcessedUpdateID = id
			}
		}
	}
}

func handlePrivacy(ctx *telegram.Context) error {
	var sb strings.Builder

	if err := templates.ExecuteTemplate(&sb, "privacy.tmpl", ctx.UserID()); err != nil {
		return fmt.Errorf("render privacy policy: %w", err)
	}

	return ctx.ReplyMarkdown(
		sb.String(),
		option.Keyboard(telegram.KeyboardPrivacy()),
	)
}
