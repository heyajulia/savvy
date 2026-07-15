package telegram

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/heyajulia/savvy/internal/telegram/chatid"
)

var (
	reaction       = marshal([]map[string]string{{"type": "emoji", "emoji": "⚡"}})
	allowedUpdates = marshal([]string{"message", "callback_query"})
)

// rawUpdate is the wire-format representation of a Telegram update.
type rawUpdate struct {
	ID            int64           `json:"update_id"`
	Message       *rawMessage     `json:"message"`
	CallbackQuery *rawCallbackQuery `json:"callback_query"`
}

type rawUser struct {
	ID chatid.ChatID `json:"id"`
}

type rawMessage struct {
	ID   int64    `json:"message_id"`
	From rawUser  `json:"from"`
	Text *string  `json:"text"`
}

type rawCallbackQuery struct {
	ID      string      `json:"id"`
	From    rawUser     `json:"from"`
	Message rawMessage  `json:"message"`
	Data    string      `json:"data"`
}

func sendRequest[T any](token, method string, parameters url.Values) (*T, error) {
	type result[T any] struct {
		OK          bool    `json:"ok"`
		Description *string `json:"description"`
		Result      T       `json:"result"`
	}

	resp, err := http.PostForm(fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, method), parameters)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	var r result[T]

	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("decode body: %w", err)
	}

	if !r.OK {
		description := "no description"
		if r.Description != nil {
			description = *r.Description
		}

		return nil, fmt.Errorf("not ok: %s", description)
	}

	return &r.Result, nil
}
