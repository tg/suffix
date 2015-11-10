package suffix

import (
	"strings"
	"testing"
)

func TestSet_Find(t *testing.T) {
	var set Set

	if set.Has("test.com") {
		t.Error("test.com belongs to empty set")
	}

	set.Add("redtube.com")
	set.Add("bbc.co.uk")
	set.Add("a.b.c.d")
	set.Add(".arpa..")

	cases := []struct {
		name   string
		suffix string
	}{
		{"redtube.com", "redtube.com"},
		{"gang.redtube.com", "redtube.com"},
		{"bang.gang.redtube.com", "redtube.com"},

		{"a.b.c.d", "a.b.c.d"},
		{".a.b.c.d", "a.b.c.d"},

		{"arpa", "arpa"},
		{"in.arpa", "arpa"},

		{"a.a.c.d", ""},
		{"pinktube.com", ""},
		{"edtube.com", ""},
		{"com", ""},
		{"bbc", ""},
		{"co.uk", ""},
		{"bbc.co", ""},
	}

	for _, c := range cases {
		suffix := set.Match(c.name)
		if suffix != c.suffix {
			t.Errorf("%s expected suffix %q, got %q", c.name, c.suffix, suffix)
		}
		belongs := (suffix != "")
		if b := set.Matches(c.name); b != belongs {
			t.Error(c.name, b, " != ", belongs)
		}
	}
}

func TestSet_ReadFrom(t *testing.T) {
	var set Set

	data := `// This is a file full of suffixes
two.girls
# comment
  one.cup  `

	expected := []string{"two.girls", "one.cup"}

	set.ReadFrom(strings.NewReader(data))

	if set.Len() != len(expected) {
		t.Fatalf("Expected %d suffixes, got %d", len(expected), set.Len())
	}

	for _, e := range expected {
		if !set.Has(e) {
			t.Errorf("Set doesn't have %q", e)
		}
	}
}
