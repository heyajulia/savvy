package telegram

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/heyajulia/savvy/internal/mustjson"
	"github.com/heyajulia/savvy/internal/telegram/chatid"
)

var (
	reaction       = mustjson.Encode([]map[string]string{{"type": "emoji", "emoji": "⚡"}})
	allowedUpdates = mustjson.Encode([]string{"message", "callback_query"})
)

type update struct {
	ID            float64        `json:"update_id"`
	Message       *message       `json:"message"`
	CallbackQuery *callbackQuery `json:"callback_query"`
}

func (u update) IsMessage() bool {
	return u.Message != nil
}

func (u update) IsCallbackQuery() bool {
	return u.CallbackQuery != nil
}

func (u update) UserID() chatid.ChatID {
	if u.IsMessage() {
		return u.Message.From.ID
	}

	return u.CallbackQuery.From.ID
}

type user struct {
	ID chatid.ChatID `json:"id"`
}

type message struct {
	ID   float64 `json:"message_id"`
	From user    `json:"from"`
	Text *string `json:"text"`
}

type callbackQuery struct {
	ID      string  `json:"id"`
	From    user    `json:"from"`
	Message message `json:"message"`
	Data    string  `json:"data"`
}

type client struct {
	token string
}

func NewClient(token string) *client {
	return &client{token}
}

func (c *client) GetUpdates(offset int64) ([]update, error) {
	updates, err := sendRequest[[]update](c.token, "getUpdates", url.Values{
		"offset":          {strconv.FormatInt(offset, 10)},
		"timeout":         {"60"},
		"allowed_updates": {allowedUpdates},
	})
	if err != nil {
		return nil, fmt.Errorf("telegram: getUpdates: %w", err)
	}

	return *updates, nil
}

func (c *client) DeleteMessage(chatID chatid.ChatID, messageID int64) error {
	return c.fireOff("deleteMessage", url.Values{
		"chat_id":    {chatID.String()},
		"message_id": {strconv.FormatInt(messageID, 10)},
	})
}

func (c *client) AnswerCallbackQuery(id string) error {
	return c.fireOff("answerCallbackQuery", url.Values{
		"callback_query_id": {id},
	})
}

func (c *client) SetMessageReaction(chatID chatid.ChatID, messageID int64) error {
	return c.fireOff("setMessageReaction", url.Values{
		"chat_id":    {chatID.String()},
		"message_id": {strconv.FormatInt(messageID, 10)},
		"is_big":     {"true"},
		"reaction":   {reaction},
	})
}

func (c *client) SendMessage(chatID chatid.ChatID, text string, parseMode parseMode, keyboard keyboard) (*message, error) {
	parameters := url.Values{
		"chat_id": {chatID.String()},
		"text":    {text},
	}

	if parseMode != ParseModeDefault {
		parameters.Add("parse_mode", parseMode.String())
	}

	if keyboard != KeyboardNone {
		parameters.Add("reply_markup", keyboard.String())
	}

	message, err := sendRequest[message](c.token, "sendMessage", parameters)
	if err != nil {
		return nil, fmt.Errorf("telegram: sendMessage: %w", err)
	}

	return message, nil
}

func (c *client) fireOff(method string, parameters url.Values) error {
	if _, err := sendRequest[any](c.token, method, parameters); err != nil {
		return fmt.Errorf("telegram: %s: %w", method, err)
	}

	return nil
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
