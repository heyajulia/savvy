package internal

import "time"

type greeting struct {
	Hello   string
	Goodbye string
}

var Afternoon = greeting{Hello: "Goedemiddag! ☀️", Goodbye: "Fijne dag verder!"}
var Evening = greeting{Hello: "Goedenavond! 🌙", Goodbye: "Geniet van je avond!"}

func GetGreeting() greeting {
	loc, _ := time.LoadLocation("Europe/Amsterdam")
	now := time.Now().In(loc)
	hour := now.Hour()

	if hour < 18 {
		return Afternoon
	}

	return Evening
}
