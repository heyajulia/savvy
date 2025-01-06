package telegram

import "github.com/heyajulia/energieprijzen/internal/mustjson"

var (
	start = mustjson.Encode(map[string]any{
		"inline_keyboard": [][]map[string]string{
			{{"text": "📜 Lees hoe ik met je privacy omga", "callback_data": "privacy"}},
			{{"text": "❤️ Abonneer je op mijn kanaal", "url": "https://t.me/energieprijzen"}},
			{{"text": "🏙️ Volg me op Bluesky", "url": "https://bsky.app/profile/bot.julia.cool"}},
		},
	})
	privacy = mustjson.Encode(map[string]any{
		"inline_keyboard": [][]map[string]string{
			{{"text": "🚮 Verwijder dit bericht", "callback_data": "got_it"}},
		},
	})

	KeyboardNone    = keyboard{""}
	KeyboardStart   = keyboard{start}
	KeyboardPrivacy = keyboard{privacy}
)

type keyboard struct {
	keyboard string
}

func (k keyboard) String() string {
	return k.keyboard
}
