package internal

import "github.com/heyajulia/energieprijzen/internal/date"

func GetGreeting() (hello string, goodbye string) {
	hour := date.Amsterdam().Hour()

	if hour < 18 {
		hello = "Goedemiddag! ☀️"
		goodbye = "Fijne dag verder!"
	} else {
		hello = "Goedenavond! 🌙"
		goodbye = "Geniet van je avond!"
	}

	return
}
