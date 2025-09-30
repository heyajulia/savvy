package option

import (
	"net/url"
	"testing"
)

func TestParseModeHTML(t *testing.T) {
	values := url.Values{}

	ParseModeHTML(values)

	if got := values.Get("parse_mode"); got != "HTML" {
		t.Fatalf("expected parse_mode to be HTML, got %q", got)
	}
}

func TestParseModeMarkdown(t *testing.T) {
	values := url.Values{}

	ParseModeMarkdown(values)

	if got := values.Get("parse_mode"); got != "Markdown" {
		t.Fatalf("expected parse_mode to be Markdown, got %q", got)
	}
}

func TestKeyboard(t *testing.T) {
	values := url.Values{}
	markup := `{"keyboard":[]}` // shape doesn't matter as long as it is carried through

	opt := Keyboard(markup)
	opt(values)

	if got := values.Get("reply_markup"); got != markup {
		t.Fatalf("expected reply_markup to be %q, got %q", markup, got)
	}
}

func TestAllOptionsTogether(t *testing.T) {
	values := url.Values{}
	markup := `{"keyboard":[]}`

	ParseModeHTML(values)
	Keyboard(markup)(values)
	ParseModeMarkdown(values)

	if got := values.Get("parse_mode"); got != "Markdown" {
		t.Fatalf("expected parse_mode to be Markdown after overrides, got %q", got)
	}

	if got := values.Get("reply_markup"); got != markup {
		t.Fatalf("expected reply_markup to be %q, got %q", markup, got)
	}
}
