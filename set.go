package suffix

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"
)

type matchType uint8

const (
	matchNone matchType = iota

	// matchMore doesn't match anything directly,
	// but indicates there are more matching suffixes.
	matchMore = 1 << iota

	matchExact = 1 << iota // match exact names
	matchSub   = 1 << iota // match sub-names

	matchAll = matchExact | matchSub
)

func (m matchType) has(f matchType) bool {
	return m&f == f
}

// decodeSuffix return suffix and match type based on the pattern.
// If trailing dot is present then matchExact is set,
// leading dot yields matchSub, otherwise matchAll.
func decodeSuffix(suffix string) (string, matchType) {
	var match matchType

	if suffix[len(suffix)-1] == '.' {
		match |= matchExact
	}
	if suffix[0] == '.' {
		match |= matchSub
	}

	if match == matchNone {
		match = matchAll
	}

	return strings.Trim(suffix, "."), match
}

// encodeSuffix is opposite to decodeSuffix, appending dot if necessary.
func encodeSuffix(suffix string, match matchType) string {
	switch true {
	case match.has(matchAll):
		return suffix
	case match.has(matchExact):
		return suffix + "."
	case match.has(matchSub):
		return "." + suffix
	}

	return ""
}

// Set defines set of suffixes
type Set struct {
	names map[string]matchType
	size  int
}

// Len returns number of suffixes in Set
func (set *Set) Len() int {
	return set.size
}

// Match contains matching suffix and way it matches
type Match struct {
	Suffix string // raw suffix (domain)
	Exact  bool   // if true exact (full) match, otherwise subdomain match
}

// Add suffix to the set. If suffix starts with a dot, only values ending,
// but not equal will be matched; if suffix ends with a dot, only exact
// values will be matched. E.g.:
//   "golang.org" will match golang.org and blog.golang.org
//   ".golang.org" will match blog.golang.org, but not golang.org
//   "golang.org." will match golang.org only
//   ".golang.org." is equivalent to "golang.org"
func (set *Set) Add(suffix string) []Match {
	if len(suffix) == 0 {
		return nil
	}

	if set.names == nil {
		set.names = make(map[string]matchType)
	}

	var match matchType
	suffix, match = decodeSuffix(suffix)

	// Prepare resulting matches
	res := make([]Match, 0, 2)
	if match&matchExact != 0 {
		res = append(res, Match{suffix, true})
	}
	if match&matchSub != 0 {
		res = append(res, Match{suffix, false})
	}

	// Increase size if suffix didn't match anything before
	if set.names[suffix]&matchAll == 0 {
		set.size++
	}
	set.names[suffix] |= match

	// Add all parent names to build a tree.
	// Don't need to do this for matchExact as we check them directly.
	if match != matchExact {
		for len(suffix) > 0 {
			dot := strings.IndexByte(suffix, '.')
			if dot < 0 {
				break
			}
			suffix = suffix[dot+1:]
			set.names[suffix] |= matchMore
		}
	}

	return res
}

// MatchAll calls callback for each matching suffix.
func (set *Set) MatchAll(name string, callback func(m Match) bool) {
	if len(set.names) == 0 {
		return
	}

	// Check exact match first, so we only care about parent suffixes later.
	// Also means we don't always need to track all parent suffixes in Add().
	if set.MatchesExact(name) {
		if !callback(Match{name, true}) {
			return
		}
	}

	// Check sub-matches by starting with the last label
	dot := len(name)
	for {
		dot = strings.LastIndexByte(name[:dot], '.')
		if dot < 0 {
			break
		}

		s := name[dot+1:] // extract current suffix
		m := set.names[s] // check match

		if m.has(matchSub) {
			if !callback(Match{s, false}) {
				break
			}
		}
		if !m.has(matchMore) {
			break
		}
	}
}

// Match returns the longest matching suffix.
// If nothing matches empty string is returned.
func (set *Set) Match(name string) string {
	var res string
	set.MatchAll(name, func(m Match) bool {
		res = m.Suffix
		return !m.Exact // stop on exact match, otherwise keep matching
	})
	return res
}

// Matches checks if passed name matches any suffix.
// This is potentially quicker than using Match(name) != "" as we stop
// searching after the first match.
func (set *Set) Matches(name string) bool {
	var res bool
	set.MatchAll(name, func(Match) bool {
		res = true
		return false // stop on first match
	})
	return res
}

// MatchesExact returns true if name matches exactly.
// Similar to Match(name) == name, but requires only a single lookup.
func (set *Set) MatchesExact(name string) bool {
	return len(set.names) > 0 && set.names[name]&matchExact != 0
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
	suffs := make([]string, 0, set.Len())
	for s, m := range set.names {
		s = encodeSuffix(s, m)
		if s != "" {
			suffs = append(suffs, s)
		}
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
