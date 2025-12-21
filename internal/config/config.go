package config

import (
	"context"

	"github.com/heyajulia/savvy/internal/telegram/chatid"
	"github.com/sethvargo/go-envconfig"
)

type Configuration struct {
	Telegram telegram `env:", prefix=TG_, required"`
	Bluesky  bluesky  `env:", prefix=BS_, required"`
	Cronitor cronitor `env:", prefix=CR_"`
	StampDir string   `env:"STAMP_DIR, required"`
}

type telegram struct {
	ChatID chatid.ChatID `env:"CHAT_ID, required"`
	Token  string        `env:"TOKEN, required"`
}

type bluesky struct {
	Identifier string `env:"IDENTIFIER, required"`
	Password   string `env:"PASSWORD, required"`
}

type cronitor struct {
	URL string `env:"URL"`
}

func Read() (Configuration, error) {
	var c Configuration

	if err := envconfig.Process(context.Background(), &c); err != nil {
		return Configuration{}, err
	}

	return c, nil
}
