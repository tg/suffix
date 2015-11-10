package main

import (
	"bytes"
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

	expected := `3 . 3
4 4 4 dot.dot
.
`
	inverted := `1 1 1
2 2

5
`

	filter := Filter{
		// Match any name containg dot
		MatchField: func(n string) bool {
			return strings.Contains(n, ".")
		},
	}

	var buf bytes.Buffer
	err := filter.Run(NewSplitScanner(strings.NewReader(input), " "), &buf)
	if err != nil {
		t.Error(err)
	}

	if out := buf.String(); out != expected {
		t.Errorf("\n-- Expected:\n%s-- Got:\n%s", expected, out)
	}

	filter.MatchNone = true
	buf.Reset()
	err = filter.Run(NewSplitScanner(strings.NewReader(input), " "), &buf)
	if err != nil {
		t.Error(err)
	}
	if out := buf.String(); out != inverted {
		t.Errorf("\n-- Expected:\n%s-- Got:\n%s", inverted, out)
	}
}
