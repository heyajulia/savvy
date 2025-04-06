package internal

import (
	"net/url"
	"testing"
	"time"
)

func TestQueryParameters(t *testing.T) {
	// Load the Europe/Amsterdam timezone.
	loc, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		t.Fatalf("failed to load Europe/Amsterdam timezone: %v", err)
	}

	testCases := []struct {
		name string
		// Input time is the time when the query is made.
		inputTime      time.Time
		expectedParams url.Values
	}{
		{
			name:      "Standard day: query on March 28 for March 29",
			inputTime: time.Date(2025, time.March, 28, 12, 0, 0, 0, loc),
			// March 29 in Amsterdam (CET, UTC+1) starts at local 00:00,
			// which in UTC is 2025-03-28T23:00:00Z, and ends at 2025-03-29T22:59:59.999Z.
			expectedParams: url.Values{
				"fromDate":  {"2025-03-28T23:00:00Z"},
				"tillDate":  {"2025-03-29T22:59:59.999Z"},
				"interval":  {"4"},
				"usageType": {"1"},
				"inclBtw":   {"true"},
			},
		},
		{
			name:      "DST start day: query on March 29 for March 30",
			inputTime: time.Date(2025, time.March, 29, 12, 0, 0, 0, loc),
			// March 30 is the DST transition day.
			// Local midnight March 30 is before the DST jump (UTC+1) resulting in 2025-03-29T23:00:00Z,
			// while the local end-of-day (23:59:59.999) after the jump (UTC+2) is 2025-03-30T21:59:59.999Z.
			expectedParams: url.Values{
				"fromDate":  {"2025-03-29T23:00:00Z"},
				"tillDate":  {"2025-03-30T21:59:59.999Z"},
				"interval":  {"4"},
				"usageType": {"1"},
				"inclBtw":   {"true"},
			},
		},
		{
			name:      "Full DST day: query on March 30 for March 31",
			inputTime: time.Date(2025, time.March, 30, 12, 0, 0, 0, loc),
			// March 31 in Amsterdam is fully in DST (UTC+2):
			// Local midnight (00:00) converts to 2025-03-30T22:00:00Z,
			// and end-of-day converts to 2025-03-31T21:59:59.999Z.
			expectedParams: url.Values{
				"fromDate":  {"2025-03-30T22:00:00Z"},
				"tillDate":  {"2025-03-31T21:59:59.999Z"},
				"interval":  {"4"},
				"usageType": {"1"},
				"inclBtw":   {"true"},
			},
		},
		{
			name:      "DST end day: query on October 25 for October 26",
			inputTime: time.Date(2025, time.October, 25, 12, 0, 0, 0, loc),
			// October 26 is the DST end day.
			// Local midnight October 26 (00:00) is still in DST (UTC+2), so in UTC that's 2025-10-25T22:00:00Z.
			// After the fallback at 03:00, the end-of-day (23:59:59.999) is in CET (UTC+1),
			// converting to 2025-10-26T22:59:59.999Z.
			expectedParams: url.Values{
				"fromDate":  {"2025-10-25T22:00:00Z"},
				"tillDate":  {"2025-10-26T22:59:59.999Z"},
				"interval":  {"4"},
				"usageType": {"1"},
				"inclBtw":   {"true"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params, err := QueryParameters(tc.inputTime)
			if err != nil {
				t.Fatal(err)
			}

			for key, expectedValues := range tc.expectedParams {
				actualValues, ok := params[key]
				if !ok {
					t.Errorf("key %q not found in parameters", key)
					continue
				}
				if len(actualValues) != 1 {
					t.Errorf("expected one value for key %q, got %d", key, len(actualValues))
					continue
				}
				if actualValues[0] != expectedValues[0] {
					t.Errorf("for key %q, expected %q but got %q", key, expectedValues[0], actualValues[0])
				}
			}
		})
	}
}
