package config

import (
	"encoding/json"
	"fmt"
	"io"
)

// Verify interface compliance.
var (
	_ valider = (*telegram)(nil)
	_ valider = (*bluesky)(nil)
	_ valider = (*cronitor)(nil)
	_ valider = (*Configuration)(nil)
)

type valider interface {
	Valid() error
}

type Configuration struct {
	Telegram telegram `json:"telegram"`
	Bluesky  bluesky  `json:"bluesky"`
	Cronitor cronitor `json:"cronitor"`
}

func (c Configuration) Valid() error {
	if err := c.Telegram.Valid(); err != nil {
		return fmt.Errorf("config: telegram: %w", err)
	}

	if err := c.Bluesky.Valid(); err != nil {
		return fmt.Errorf("config: bluesky: %w", err)
	}

	if err := c.Cronitor.Valid(); err != nil {
		return fmt.Errorf("config: cronitor: %w", err)
	}

	return nil
}

func Read(r io.Reader) (Configuration, error) {
	var c Configuration

	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&c); err != nil {
		return Configuration{}, fmt.Errorf("config: decode: %w", err)
	}

	if err := c.Valid(); err != nil {
		return Configuration{}, err
	}

	return c, nil
}
