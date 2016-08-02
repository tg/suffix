package main

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

func TestSplitFilter(t *testing.T) {
	input := `1 1 1
2 2
3 . 3
4 4 4 dot.dot

5
.
`

	cases := []struct {
		Filter   *Filter
		Expected string
	}{
		{&Filter{}, "3 . 3\n4 4 4 dot.dot\n.\n"},
		{&Filter{MatchNone: true}, "1 1 1\n2 2\n\n5\n"},
		{&Filter{OnlyMatch: true}, ".\n.dot\n.\n"},
		{&Filter{OnlyField: true}, ".\ndot.dot\n.\n"},
	}

	matchField := func(field string) string {
		return regexp.MustCompile(`[.]\w*`).FindString(field)
	}

	for _, c := range cases {
		filter := c.Filter
		filter.MatchField = matchField

		var buf bytes.Buffer
		err := filter.Run(NewSplitScanner(strings.NewReader(input), " "), &buf)
		if err != nil {
			t.Error(err)
		}

		if out := buf.String(); out != c.Expected {
			t.Errorf("\n-- Expected:\n%s-- Got:\n%s", c.Expected, out)
		}
	}
}
