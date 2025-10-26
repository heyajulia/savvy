package main

import (
	"slices"
	"testing"
	"time"
)

func TestHourNumbersForDay(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}

	testCases := []struct {
		name     string
		day      time.Time
		count    int
		expected []int
	}{
		{
			name:     "regular day",
			day:      time.Date(2024, time.March, 15, 12, 0, 0, 0, loc),
			count:    24,
			expected: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
		},
		{
			name:  "dst start day loses hour",
			day:   time.Date(2025, time.March, 30, 12, 0, 0, 0, loc),
			count: 23,
			expected: []int{
				0, 1, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
				13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23,
			},
		},
		{
			name:  "dst end day repeats hour",
			day:   time.Date(2024, time.October, 27, 12, 0, 0, 0, loc),
			count: 25,
			expected: []int{
				0, 1, 2, 2, 3, 4, 5, 6, 7, 8, 9, 10,
				11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := hourNumbersForDay(tc.day, tc.count)
			if !slices.Equal(actual, tc.expected) {
				t.Fatalf("expected %v, got %v", tc.expected, actual)
			}
		})
	}
}

func TestFormatHourRanges(t *testing.T) {
	tests := []struct {
		name     string
		indexes  []int
		hours    []int
		expected string
	}{
		{
			name:     "deduplicates repeated hour",
			indexes:  []int{2, 3},
			hours:    []int{0, 1, 2, 2, 3},
			expected: "van 02:00 tot 02:59",
		},
		{
			name:     "handles missing hour",
			indexes:  []int{0, 1, 2},
			hours:    []int{0, 1, 3, 4},
			expected: "van 00:00 tot 01:59 en van 03:00 tot 03:59",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := formatHourRanges(tc.indexes, tc.hours)
			if actual != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, actual)
			}
		})
	}
}

