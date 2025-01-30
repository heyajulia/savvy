package config

import "errors"

var (
	errEmptyIdentifier = errors.New("empty identifier")
	errEmptyPassword   = errors.New("empty password")
)

type bluesky struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

func (b bluesky) Valid() error {
	if b.Identifier == "" {
		return errEmptyIdentifier
	}

	if b.Password == "" {
		return errEmptyPassword
	}

	return nil
}
