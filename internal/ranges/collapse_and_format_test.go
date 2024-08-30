package ranges

import "testing"

func TestCollapseAndFormat(t *testing.T) {
	tests := []struct {
		values []int
		want   string
	}{
		{[]int{}, ""},
		{[]int{1}, "van 01:00 tot 01:59"},
		{[]int{1, 2, 3}, "van 01:00 tot 03:59"},
		{[]int{1, 2, 3, 5, 6, 7}, "van 01:00 tot 03:59 en van 05:00 tot 07:59"},
		{[]int{1, 2, 3, 5, 6, 7, 9, 10, 11}, "van 01:00 tot 03:59, van 05:00 tot 07:59 en van 09:00 tot 11:59"},
		{[]int{1, 2, 4, 6}, "van 01:00 tot 02:59, van 04:00 tot 04:59 en van 06:00 tot 06:59"},
		{[]int{1, 3, 5}, "van 01:00 tot 01:59, van 03:00 tot 03:59 en van 05:00 tot 05:59"},
		{[]int{1, 2, 10, 11}, "van 01:00 tot 02:59 en van 10:00 tot 11:59"},
		{[]int{1, 2, 3, 5, 6, 7, 9, 11, 12, 13, 15, 17, 19, 21, 22, 23}, "van 01:00 tot 03:59, van 05:00 tot 07:59, van 09:00 tot 09:59, van 11:00 tot 13:59, van 15:00 tot 15:59, van 17:00 tot 17:59, van 19:00 tot 19:59 en van 21:00 tot 23:59"},
	}

	for _, test := range tests {
		actual := CollapseAndFormat(test.values)
		if actual != test.want {
			t.Errorf("CollapseAndFormat(%v) = %q, want %q", test.values, actual, test.want)
		}
	}
}
