package config

import (
	"fmt"
	"net/url"
)

type cronitor struct {
	TelemetryURL string `json:"telemetry_url"`
}

func (c cronitor) Valid() error {
	if _, err := url.Parse(c.TelemetryURL); err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}

	return nil
}
