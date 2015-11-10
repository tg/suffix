package main

import (
	"errors"
	"fmt"
	"io"
)

// Filter reads input and prints matching lines to output
type Filter struct {
	// MatchField should return true if field matches the filter
	MatchField func(string) bool

	// Normally filter matches whole line if any field matches.
	// If MatchNone is set than no field can match in order to match the
	// whole line. This is usefull to implement inverted match.
	MatchNone bool
}

// Run scans input and print matching lines to output
func (f *Filter) Run(s LineScanner, w io.Writer) error {
	match := f.MatchField
	if match == nil {
		return errors.New("No match function defined, wouldn't write anything")
	}

	for s.Scan() {
		matched := false
		for _, field := range s.Fields() {
			if match(field) {
				matched = true
				break
			}
		}
		if !matched == f.MatchNone {
			fmt.Fprintln(w, s.Text())
		}
	}
	return s.Err()
}
