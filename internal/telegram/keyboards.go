package telegram

var (
	KeyboardStart = marshal(map[string]any{
		"inline_keyboard": [][]map[string]string{
			{{"text": "📜 Lees hoe ik met je privacy omga", "callback_data": "privacy"}},
			{{"text": "❤️ Abonneer je op mijn kanaal", "url": "https://t.me/energieprijzen"}},
			{{"text": "🏙️ Volg me op Bluesky", "url": "https://bsky.app/profile/bot.julia.cool"}},
		},
	})

	KeyboardPrivacy = marshal(map[string]any{
		"inline_keyboard": [][]map[string]string{
			{{"text": "🚮 Verwijder dit bericht", "callback_data": "got_it"}},
		},
	})
)
