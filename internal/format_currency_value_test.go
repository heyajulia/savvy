package internal

import (
	"math"
	"testing"
)

func TestPositiveZero(t *testing.T) {
	actual := FormatCurrencyValue(0)
	if actual != "€\u00a00,00" {
		t.Errorf("FormatCurrencyValue(0) = %q, want %q", actual, "€\u00a00,00")
	}
}

func TestNegativeZero(t *testing.T) {
	negativeZero := math.Copysign(0, -1)
	actual := FormatCurrencyValue(negativeZero)
	if actual != "€\u00a00,00" {
		t.Errorf("FormatCurrencyValue(-0) = %q, want %q", actual, "€\u00a00,00")
	}
}

func TestPositive(t *testing.T) {
	actual := FormatCurrencyValue(1)
	if actual != "€\u00a01,00" {
		t.Errorf("FormatCurrencyValue(1) = %q, want %q", actual, "€\u00a01,00")
	}
}

func TestNegative(t *testing.T) {
	actual := FormatCurrencyValue(-1)
	if actual != "€\u00a0-1,00" {
		t.Errorf("FormatCurrencyValue(-1) = %q, want %q", actual, "€\u00a0-1,00")
	}
}
