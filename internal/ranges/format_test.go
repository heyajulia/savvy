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
		{[]Range{New(1, 3), New(5, 7), Single(9), New(11, 13), Single(15), Single(17), New(19, 20), Single(21), New(22, 23)}, "van 01:00 tot 03:59, van 05:00 tot 07:59, van 09:00 tot 09:59, van 11:00 tot 13:59, van 15:00 tot 15:59, van 17:00 tot 17:59, van 19:00 tot 20:59, van 21:00 tot 21:59 en van 22:00 tot 23:59"}}

	for _, test := range tests {
		actual := Format(test.ranges)
		if actual != test.want {
			t.Errorf("Format(%v) = %q, want %q", test.ranges, actual, test.want)
		}
	}
}

func BenchmarkFormat(b *testing.B) {
	r := []Range{New(1, 3), New(5, 7), Single(9)}
	for range b.N {
		_ = Format(r)
	}
}

func BenchmarkFormatMore(b *testing.B) {
	r := []Range{New(1, 3), New(5, 7), Single(9), New(11, 13), Single(15), Single(17), New(19, 20), Single(21), New(22, 23)}
	for range b.N {
		_ = Format(r)
	}
}
