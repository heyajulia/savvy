package ranges

import (
	"slices"
	"testing"
)

func TestCollapse(t *testing.T) {
	tests := []struct {
		values []int
		want   []Range
	}{
		{[]int{}, []Range{}},
		{nil, []Range{}},
		{[]int{}, nil},
		{[]int{1}, []Range{Single(1)}},
		{[]int{1, 2, 3}, []Range{New(1, 3)}},
		{[]int{1, 2, 3, 5, 6, 7}, []Range{New(1, 3), New(5, 7)}},
		{[]int{1, 2, 3, 5, 6, 7, 9, 10, 11}, []Range{New(1, 3), New(5, 7), New(9, 11)}},
		{[]int{1, 2, 4, 6}, []Range{New(1, 2), Single(4), Single(6)}},
		{[]int{1, 3, 5}, []Range{Single(1), Single(3), Single(5)}},
		{[]int{1, 2, 10, 11}, []Range{New(1, 2), New(10, 11)}},
		{[]int{1, 2, 3, 5, 6, 7, 9, 11, 12, 13, 15, 17, 19, 21, 22, 23, 25}, []Range{New(1, 3), New(5, 7), Single(9), New(11, 13), Single(15), Single(17), Single(19), New(21, 23), Single(25)}},
	}

	for _, test := range tests {
		actual := Collapse(test.values)
		if !slices.Equal(actual, test.want) {
			t.Errorf("Collapse(%v) == %v, want %v", test.values, actual, test.want)
		}
	}
}
