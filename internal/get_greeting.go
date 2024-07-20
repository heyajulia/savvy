package internal

import "github.com/heyajulia/energieprijzen/internal/datetime"

func GetGreeting() (hello, goodbye string) {
	hour := datetime.Amsterdam().Hour()

	if hour < 18 {
		hello = "Goedemiddag! ☀️"
		goodbye = "Fijne dag verder!"
	} else {
		hello = "Goedenavond! 🌙"
		goodbye = "Geniet van je avond!"
	}

	return
}
