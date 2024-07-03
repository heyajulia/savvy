package internal

import (
	"fmt"
	"math"
	"strings"
)

func FormatCurrencyValue(value float64) string {
	v := math.Round(value*100) / 100
	if v == 0 && math.Signbit(v) { // is v negative zero?
		v = 0
	}
	return strings.Replace(fmt.Sprintf("€\u00a0%.2f", v), ".", ",", 1)
}
