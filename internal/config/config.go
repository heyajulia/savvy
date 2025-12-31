package config

import (
	"context"
	"fmt"

	"github.com/heyajulia/savvy/internal/telegram/chatid"
	"github.com/sethvargo/go-envconfig"
)

// TelegramBase contains Telegram configuration shared by both serve and report.
type TelegramBase struct {
	Token       string `env:"TOKEN, required"`
	ChannelName string `env:"CHANNEL_NAME, default=energieprijzen"`
}

// TelegramReport extends TelegramBase with fields only needed by report.
type TelegramReport struct {
	TelegramBase
	ChatID chatid.ChatID `env:"CHAT_ID, required"`
}

// BlueskyBase contains Bluesky configuration shared by both serve and report.
type BlueskyBase struct {
	Identifier string `env:"IDENTIFIER, required"`
}

// BlueskyReport extends BlueskyBase with fields only needed by report.
type BlueskyReport struct {
	BlueskyBase
	Password string `env:"PASSWORD, required"`
}

// Cronitor contains optional Cronitor monitoring configuration.
type Cronitor struct {
	URL string `env:"URL"`
}

// Serve contains configuration for the serve binary.
type Serve struct {
	Telegram TelegramBase `env:", prefix=TG_"`
	Bluesky  BlueskyBase  `env:", prefix=BS_"`
}

// Report contains configuration for the report binary.
type Report struct {
	Telegram TelegramReport `env:", prefix=TG_"`
	Bluesky  BlueskyReport  `env:", prefix=BS_"`
	Cronitor Cronitor       `env:", prefix=CR_"`
	StampDir string         `env:"STAMP_DIR, required"`
}

// Read reads configuration from environment variables into the given type.
func Read[T any]() (T, error) {
	var c T

	if err := envconfig.Process(context.Background(), &c); err != nil {
		var zero T
		return zero, fmt.Errorf("config: process config: %w", err)
	}

	return c, nil
}
