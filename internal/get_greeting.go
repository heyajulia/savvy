package internal

import "time"

type greeting struct {
	Hello   string
	Goodbye string
}

var Afternoon = greeting{Hello: "Goedemiddag! ☀️", Goodbye: "Fijne dag verder!"}
var Evening = greeting{Hello: "Goedenavond! 🌙", Goodbye: "Geniet van je avond!"}

func GetGreeting() greeting {
	amsterdam, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		panic(err)
	}

	now := time.Now().In(amsterdam)
	hour := now.Hour()

	if hour < 18 {
		return Afternoon
	}

	return Evening
}
