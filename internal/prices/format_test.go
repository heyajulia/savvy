package prices

import (
	"math"
	"testing"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		name string
		give float64
		want string
	}{
		{
			name: "0",
			give: 0,
			want: "€\u00a00,00",
		},
		{
			name: "-0",
			give: math.Copysign(0, -1),
			want: "€\u00a00,00",
		},
		{
			name: "1",
			give: 1,
			want: "€\u00a01,00",
		},
		{
			name: "-1",
			give: -1,
			want: "€\u00a0-1,00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := Format(tt.give)
			if actual != tt.want {
				t.Errorf("got %q, want %q", actual, tt.want)
			}
		})
	}
}
