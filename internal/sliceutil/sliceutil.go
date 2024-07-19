package sliceutil

import (
	"math"
	"slices"
)

// Min returns the minimum value in the given slice of float64 values. If values is empty, it returns positive infinity.
func Min(values []float64) float64 {
	if len(values) < 1 {
		return math.Inf(1)
	}

	return slices.Min(values)
}

// Max returns the maximum value in the given slice of float64 values. If values is empty, it returns negative infinity.
func Max(values []float64) float64 {
	if len(values) < 1 {
		return math.Inf(-1)
	}

	return slices.Max(values)
}
