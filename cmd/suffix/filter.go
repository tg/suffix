package main

import (
	"errors"
	"fmt"
	"io"
)

// Filter reads input and prints matching lines to output
type Filter struct {
	// MatchField should return matching suffix if any
	MatchField func(string) string

	// Normally filter matches whole line if any field matches.
	// If MatchNone is set than no field can match in order to match the
	// whole line. This is usefull to implement inverted match.
	MatchNone bool

	// OnlyField makes filter printing matching field instead of the whole line.
	OnlyField bool

	// OnlyMatched makes filter printing returned match instead of the whole line.
	OnlyMatch bool
}

// Run scans input and print matching lines to output
func (f *Filter) Run(s LineScanner, w io.Writer) error {
	matchField := f.MatchField
	if matchField == nil {
		return errors.New("No match function defined, wouldn't write anything")
	}

	for s.Scan() {
		var field, match string
		for _, field = range s.Fields() {
			if m := matchField(field); m != "" {
				match = m
				break
			}
		}
		if (match == "") == f.MatchNone {
			switch {
			case f.OnlyMatch:
				fmt.Fprintln(w, match)
			case f.OnlyField:
				fmt.Fprintln(w, field)
			default:
				fmt.Fprintln(w, s.Text())
			}
		}
	}
	return s.Err()
}
