package suffix

import (
	"bufio"
	"io"
	"strings"
)

// Set defines set of suffixes
type Set struct {
	names     map[string]struct{}
	maxLabels int
}

// Len returns number of entries in Set
func (set *Set) Len() int {
	return len(set.names)
}

// Add suffix to the set
func (set *Set) Add(suffix string) {
	if set.names == nil {
		set.names = make(map[string]struct{})
	}

	suffix = strings.Trim(suffix, ".")
	set.names[suffix] = struct{}{}
	// Find max number of lables
	// TODO: handle double dot (*..*)
	labels := strings.Count(suffix, ".") + 1
	if labels > set.maxLabels {
		set.maxLabels = labels
	}
}

// Has returns true iff suffix was added to set.
func (set *Set) Has(suffix string) bool {
	_, ok := set.names[suffix]
	return ok
}

// Match returns a matching suffix.
// If nothing matches empty string is returned.
func (set *Set) Match(name string) string {
	if len(set.names) > 0 {
		dot := len(name)
		for n := 0; n < set.maxLabels && dot >= 0; n++ {
			dot = strings.LastIndex(name[:dot], ".")
			suffix := name[dot+1:]
			if _, ok := set.names[suffix]; ok {
				return suffix
			}
		}
	}
	return ""
}

// Matches checks if passed name matches any suffix.
// Equivalent to Match(name) != ""
func (set *Set) Matches(name string) bool {
	return set.Match(name) != ""
}

// ReadFrom reads set from the stream. Each non-empty line of stream is
// considered a suffix, except from lines beginning with '#' or '//', which
// are treated as comments and skipped.
func (set *Set) ReadFrom(r io.Reader) (n int64, err error) {
	cnt := &counter{}
	scanner := bufio.NewScanner(io.TeeReader(r, cnt))
	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), " \t")
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		set.Add(line)
	}

	return cnt.N, scanner.Err()
}

type counter struct {
	N int64
}

func (c *counter) Write(p []byte) (n int, err error) {
	c.N += int64(len(p))
	return len(p), nil
}
