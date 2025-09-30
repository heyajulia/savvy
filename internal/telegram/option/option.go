package option

import "net/url"

type Option func(url.Values)

func ParseModeHTML(v url.Values) {
	v.Set("parse_mode", "HTML")
}

func ParseModeMarkdown(v url.Values) {
	v.Set("parse_mode", "Markdown")
}

func Keyboard(keyboard string) Option {
	return func(v url.Values) {
		v.Set("reply_markup", keyboard)
	}
}
