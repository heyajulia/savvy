package ranges

import (
	"slices"
	"strings"
)

const digits string = "000102030405060708091011121314151617181920212223"

type Range struct {
	start, end int
}

func New(start, end int) Range {
	if start > end {
		panic("New: start must be less than or equal to end")
	}

	if start < 0 || end < 0 {
		panic("New: start and end must be positive")
	}

	if start > 23 || end > 23 {
		panic("New: start and end must be less than or equal to 23")
	}

	return Range{start: start, end: end}
}

func Single(value int) Range {
	return New(value, value)
}

func Collapse(values []int) []Range {
	var ranges []Range

	if len(values) == 0 {
		return ranges
	}

	slices.Sort(values)

	start := values[0]
	end := values[0]

	for i := 1; i < len(values); i++ {
		value := values[i]

		if value != end+1 {
			ranges = append(ranges, New(start, end))
			start = value
		}

		end = value
	}

	ranges = append(ranges, New(start, end))

	return ranges
}

// Format formats the given ranges as a human-readable string of start and end hours.
func Format(ranges []Range) string {
	var sb strings.Builder
	sb.Grow(len(ranges) * 20)

	for i, r := range ranges {
		if i > 0 {
			if i == len(ranges)-1 {
				sb.WriteString(" en ")
			} else {
				sb.WriteString(", ")
			}
		}

		sb.WriteString("van ")
		writeHour(&sb, r.start)
		sb.WriteString(":00 tot ")
		writeHour(&sb, r.end)
		sb.WriteString(":59")
	}

	return sb.String()
}

func CollapseAndFormat(values []int) string {
	return Format(Collapse(values))
}

func writeHour(sb *strings.Builder, hour int) {
	i := hour * 2
	sb.WriteString(digits[i : i+2])
}
