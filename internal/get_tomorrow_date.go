package internal

import (
	"strings"
	"time"
)

func GetTomorrowDate() string {
	layout := "Monday _2 January 2006"
	replacer := strings.NewReplacer(
		"Monday", "maandag",
		"Tuesday", "dinsdag",
		"Wednesday", "woensdag",
		"Thursday", "donderdag",
		"Friday", "vrijdag",
		"Saturday", "zaterdag",
		"Sunday", "zondag",
		"January", "januari",
		"February", "februari",
		"March", "maart",
		"April", "april",
		"May", "mei",
		"June", "juni",
		"July", "juli",
		"August", "augustus",
		"September", "september",
		"October", "oktober",
		"November", "november",
		"December", "december",
	)

	return replacer.Replace(time.Now().AddDate(0, 0, 1).Format(layout))
}
