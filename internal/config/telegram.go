package config

import (
	"errors"
	"strings"

	"github.com/heyajulia/energieprijzen/internal/telegram/chatid"
)

var (
	errInvalidToken  = errors.New("invalid token")
	errInvalidChatID = errors.New("invalid chat id")
)

type telegram struct {
	Token  string        `json:"token"`
	ChatID chatid.ChatID `json:"chat_id"`
}

func (t telegram) Valid() error {
	if t.Token == "" {
		return errInvalidToken
	}

	switch t.ChatID.Kind() {
	case chatid.KindID:
		return nil
	case chatid.KindUsername:
		if s := t.ChatID.String(); len(s) > 1 && strings.HasPrefix(s, "@") {
			return nil
		}

		return errInvalidChatID
	}

	return nil
}
