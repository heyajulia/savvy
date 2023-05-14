package ranges

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/heyajulia/energieprijzen/internal/fp"
)

var commaRegex = regexp.MustCompile(`,([^,]*)$`)

type Range struct {
	start, end int
}

func New(start int, end int) Range {
	if start > end {
		panic("ranges.New: start must be less than or equal to end")
	}

	if start < 0 || end < 0 {
		panic("ranges.New: start and end must be positive")
	}

	return Range{start: start, end: end}
}

func Single(value int) Range {
	return New(value, value)
}

func Collapse(values []int) []Range {
	ranges := []Range{}

	if len(values) == 0 {
		return ranges
	}

	sort.Ints(values)

	start := values[0]
	end := values[0]

	for i := 1; i < len(values); i++ {
		value := values[i]

		if value == end+1 {
			end = value
		} else {
			ranges = append(ranges, New(start, end))
			start = value
			end = value
		}
	}

	ranges = append(ranges, New(start, end))

	return ranges
}

func Format(ranges []Range) string {
	rs := fp.Map(func(r Range) string {
		return fmt.Sprintf("van %02d:00 tot %02d:59", r.start, r.end)
	}, ranges)

	s := strings.Join(rs, ", ")

	return commaRegex.ReplaceAllString(s, " en$1")
}

func CollapseAndFormat(values []int) string {
	return Format(Collapse(values))
}
