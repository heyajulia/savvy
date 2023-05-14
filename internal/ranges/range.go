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
		sverb := "%02d"
		everb := "%02d"

		// If start or end is negative, the minus sign already takes up one character, so we need to add one to the verb
		// I don't know if this is worth guarding against since negative hours should never occur in practice

		if r.start < 0 {
			sverb = "%03d"
		}

		if r.end < 0 {
			everb = "%03d"
		}

		return fmt.Sprintf(fmt.Sprintf("van %s:00 tot %s:59", sverb, everb), r.start, r.end)
	}, ranges)

	s := strings.Join(rs, ", ")

	return commaRegex.ReplaceAllString(s, " en$1")
}

func CollapseAndFormat(values []int) string {
	return Format(Collapse(values))
}
