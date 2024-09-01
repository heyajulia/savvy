package main

type hourly struct {
	Emoji, PaddedHour, FormattedPrice string
}

type templateData struct {
	Hello, Goodbye                 string
	TomorrowDate                   string
	AverageFormatted, AverageHours string
	HighFormatted, HighHours       string
	LowFormatted, LowHours         string
	Hourly                         []hourly
}
