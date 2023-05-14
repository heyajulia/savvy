package internal

func GetPriceEmoji(price float64, average float64) string {
	switch {
	case price == 0:
		return "🆓"
	case price < 0:
		return "💶"
	case price <= average:
		return "✅"
	default:
		return "❌"
	}
}
