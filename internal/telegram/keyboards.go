package telegram

import "fmt"

var keyboardPrivacy = marshal(map[string]any{
	"inline_keyboard": [][]map[string]string{
		{{"text": "🚮 Verwijder dit bericht", "callback_data": "got_it"}},
	},
})

func KeyboardPrivacy() string {
	return keyboardPrivacy
}

func KeyboardStart(channelName, blueskyIdentifier string) string {
	return marshal(map[string]any{
		"inline_keyboard": [][]map[string]string{
			{{"text": "📜 Lees hoe ik met je privacy omga", "callback_data": "privacy"}},
			{{"text": "❤️ Abonneer je op mijn kanaal", "url": fmt.Sprintf("https://t.me/%s", channelName)}},
			{{"text": "🏙️ Volg me op Bluesky", "url": fmt.Sprintf("https://bsky.app/profile/%s", blueskyIdentifier)}},
		},
	})
}
