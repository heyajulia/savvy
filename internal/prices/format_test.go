package prices

import (
	"math"
	"testing"
)

func TestPositiveZero(t *testing.T) {
	actual := Format(0)
	if actual != "€\u00a00,00" {
		t.Errorf("Format(0) = %q, want %q", actual, "€\u00a00,00")
	}
}

func TestNegativeZero(t *testing.T) {
	negativeZero := math.Copysign(0, -1)
	actual := Format(negativeZero)
	if actual != "€\u00a00,00" {
		t.Errorf("Format(-0) = %q, want %q", actual, "€\u00a00,00")
	}
}

func TestPositive(t *testing.T) {
	actual := Format(1)
	if actual != "€\u00a01,00" {
		t.Errorf("Format(1) = %q, want %q", actual, "€\u00a01,00")
	}
}

func TestNegative(t *testing.T) {
	actual := Format(-1)
	if actual != "€\u00a0-1,00" {
		t.Errorf("Format(-1) = %q, want %q", actual, "€\u00a0-1,00")
	}
}
