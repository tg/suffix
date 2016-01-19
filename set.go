package suffix

import (
	"bufio"
	"fmt"
	"io"
	"sort"
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

// Match returns the longest matching suffix.
// If nothing matches empty string is returned.
func (set *Set) Match(name string) string {
	if len(set.names) == 0 {
		return ""
	}

	// Shrink to longest suffix
	dot := len(name)
	for n := set.maxLabels; n > 0 && dot > 0; n-- {
		dot = strings.LastIndexByte(name[:dot], '.')
	}
	s := name[dot+1:]

	// Find matching suffix
	for len(s) > 0 {
		if _, ok := set.names[s]; ok {
			return s
		}
		dot := strings.IndexByte(s, '.')
		if dot < 0 {
			return ""
		}
		s = s[dot+1:]
	}

	return ""
}

// Matches checks if passed name matches any suffix.
// Equivalent to Match(name) != ""
func (set *Set) Matches(name string) bool {
	return set.Match(name) != ""
}

// Split splits name into prefix and suffix where suffix is longest matching
// suffix from the set. If no suffix matches empty strings are returned.
func (set *Set) Split(name string) (pre string, suf string) {
	suf = set.Match(name)
	if suf != "" && len(name) > len(suf) {
		pre = name[:len(name)-len(suf)-1]
	}
	return
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

// WriteTo serialises set into the writer.
// Data is serialised in plain text, each suffix in a separate line.
// Suffixes are written in lexicographical order.
func (set *Set) WriteTo(w io.Writer) (n int64, err error) {
	suffs := make([]string, 0, len(set.names))
	for s := range set.names {
		suffs = append(suffs, s)
	}
	sort.Strings(suffs)
	c := &counter{W: w}
	for n := range suffs {
		_, err = fmt.Fprintln(c, suffs[n])
		if err != nil {
			break
		}
	}
	return c.N, err
}

type counter struct {
	W io.Writer
	N int64
}

func (c *counter) Write(p []byte) (n int, err error) {
	if c.W != nil {
		n, err = c.W.Write(p)
	} else {
		n = len(p)
	}
	c.N += int64(n)
	return
}

// PlusOne returns matching suffix plus one label from the name.
// For example if set containt 'com' and name is 'www.blog.com',
// this function would return 'blog.com'. Returned string is empty if there
// is no matching suffix in the set or an additional label is missing.
func PlusOne(set *Set, name string) string {
	pre, suf := set.Split(name)
	if suf == "" || pre == "" {
		return ""
	}
	return pre[strings.LastIndexByte(pre, '.')+1:] + "." + suf
}
