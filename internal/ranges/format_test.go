package ranges

import "testing"

func TestFormat(t *testing.T) {
	tests := []struct {
		ranges []Range
		want   string
	}{
		{[]Range{}, ""},
		{[]Range{Single(1)}, "van 01:00 tot 01:59"},
		{[]Range{New(1, 3)}, "van 01:00 tot 03:59"},
		{[]Range{New(1, 3), Single(5)}, "van 01:00 tot 03:59 en van 05:00 tot 05:59"},
		{[]Range{New(1, 3), New(5, 7)}, "van 01:00 tot 03:59 en van 05:00 tot 07:59"},
		{[]Range{New(1, 3), New(5, 7), Single(9)}, "van 01:00 tot 03:59, van 05:00 tot 07:59 en van 09:00 tot 09:59"},
	}

	for _, test := range tests {
		actual := Format(test.ranges)
		if actual != test.want {
			t.Errorf("Format(%v) = %q, want %q", test.ranges, actual, test.want)
		}
	}
}
