package telegram

var (
	ParseModeDefault  = parseMode{""}
	ParseModeHTML     = parseMode{"HTML"}
	ParseModeMarkdown = parseMode{"Markdown"}
)

type parseMode struct {
	parseMode string
}

func (p parseMode) String() string {
	return p.parseMode
}
