package main

import (
	"bufio"
	"io"
	"strings"
)

// LineScanner scans lines of input and breaks them into fields.
type LineScanner interface {
	Scan() bool       // Scan next line
	Text() string     // Text returns whole line as is was read
	Fields() []string // Fields returns fields extracted from text
	Err() error       // Err indicates read errors
}

// SplitScanner splits lines form standard input into the fields by separator.
// It extends bufio.Scanner by Fields() to implement LineScanner interface.
type SplitScanner struct {
	*bufio.Scanner
	sep    string
	column int
}

// NewSplitScanner return split scanner reading from r and separating field by sep.
func NewSplitScanner(r io.Reader, sep string) *SplitScanner {
	if sep == "" {
		sep = "\t"
	}
	return &SplitScanner{
		Scanner: bufio.NewScanner(r),
		sep:     sep,
	}
}

// Fields extracts fields from last read line using separator provided earlier
func (s *SplitScanner) Fields() []string {
	return strings.Split(s.Text(), s.sep)
}

// SingleFieldScanner wraps LineScanner and returns only one specified field.
type SingleFieldScanner struct {
	LineScanner
	N int
}

// Fields rerturns n-th field from fields returned by underlying scanner.
// Nil slice is returned if N is outside the range.
func (s *SingleFieldScanner) Fields() []string {
	fs := s.LineScanner.Fields()
	if n := s.N; n >= 0 && n < len(fs) {
		return fs[n : n+1]
	}
	return nil
}
