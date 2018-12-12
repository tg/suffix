package suffix_test

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/tg/suffix"
)

func TestSet_Match(t *testing.T) {
	var set suffix.Set

	// add suffixes; multiple times to make sure this is safe
	for n := 0; n < 3; n++ {
		set.Add("redtube.com")
		set.Add("bbc.co.uk")

		set.Add("d.")
		set.Add("a.b.c.d")
		set.Add(".c.d")
		set.Add("c.d.")
		set.Add(".b.c.d")

		set.Add(".subs")
		set.Add(".subs.com")

		set.Add("fixed.")
		set.Add("fixed.com.")

		set.Add(".both.com.") // equivalent to no edge dots
	}

	if size := set.Len(); size != 11 {
		t.Error("invalid set size: ", size)
	}

	cases := []struct {
		name   string
		suffix string
	}{
		{"redtube.com", "redtube.com"},
		{"gang.redtube.com", "redtube.com"},
		{"bang.gang.redtube.com", "redtube.com"},

		{"d", "d"},
		{"c.d", "c.d"},
		{"x.d", ""},
		{"c.d", "c.d"},
		{"b.c.d", "c.d"},
		{"a.b.c.d", "a.b.c.d"},
		{"x.b.c.d", "b.c.d"},
		{"x.x.c.d", "c.d"},
		{"x.x.x.d", ""},

		{"yellow.subs", "subs"},
		{"yellow.subs.com", "subs.com"},
		{"green.and.yellow.subs.com", "subs.com"},

		{"fixed", "fixed"},
		{"fixed.com", "fixed.com"},

		{"both.com", "both.com"},
		{"yes.both.com", "both.com"},

		// Don't match these...
		{"pinktube.com", ""},
		{"edtube.com", ""},
		{"com", ""},
		{"bbc", ""},
		{"co.uk", ""},
		{"bbc.co", ""},
		{"subs", ""},
		{"subs.com", ""},
		{"some.fixed", ""},
		{"some.fixed.com", ""},
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
	var set suffix.Set

	data := `// This is a file full of suffixes
two.girls
# comment
  one.cup  `

	expected := []string{"two.girls", "one.cup"}

	set.ReadFrom(strings.NewReader(data))

	if set.Len() != len(expected) {
		t.Fatalf("Expected %d suffixes, got %d", len(expected), set.Len())
	}
}

func TestSet_WriteTo(t *testing.T) {
	var set suffix.Set
	set.Add("com")
	set.Add("google.com")
	set.Add(".youtube.com")
	set.Add("blog.golang.org.")

	buf := &bytes.Buffer{}
	set.WriteTo(buf)

	s := buf.String()
	expected := `.youtube.com
blog.golang.org.
com
google.com
`

	if s != expected {
		t.Errorf("Expected %q, got %q", expected, s)
	}
}

func ExampleSet_ReadFrom_iana() {
	r, err := http.Get("http://data.iana.org/TLD/tlds-alpha-by-domain.txt")
	if err != nil {
		log.Fatal(err)
	}

	var tlds suffix.Set
	_, err = tlds.ReadFrom(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	r.Body.Close()

	fmt.Println(tlds.Match("MERRY.CHRISTMAS"))

	// Output:
	// CHRISTMAS
}

func ExampleSet_Split() {
	var set suffix.Set
	set.Add("com")
	set.Add("blogspot.com")

	fmt.Println(set.Split("bob.blogspot.com"))
	fmt.Println(set.Split("blogspot.com"))

	// Output:
	// bob blogspot.com
	//  blogspot.com
}

func Example_plusOne() {
	var set suffix.Set
	set.Add("com")
	set.Add("blogspot.com")

	fmt.Println(suffix.PlusOne(&set, "bob.dylan.blogspot.com"))
	fmt.Println(suffix.PlusOne(&set, "blogspot.com"))

	// Output:
	// dylan.blogspot.com
	//
}

// Example_map shows how to create a mapping for suffixes added to Set.
func Example_map() {
	var set suffix.Set

	// tags will keep a list of tags for each possible match
	tags := make(map[suffix.Match][]string)

	// set up rules and add to set
	rules := []struct {
		Name string
		Tag  string
	}{
		// will match google.com and all subdomains
		{"google.com", "google"},
		// will match google.com exactly
		{"google.com.", "toplevel"},
		// will match only subdomains *.google.com
		{".google.com", "subdomain"},
		// will match api.google.com exactly
		{"www.google.com.", "websearch"},
	}
	for _, rule := range rules {
		for _, match := range set.Add(rule.Name) {
			// create mapping between each match;
			// matches are not unique between names, so we need to append.
			tags[match] = append(tags[match], rule.Tag)
		}
	}

	// get all tags for www.google.com
	set.MatchAll("www.google.com", func(m suffix.Match) bool {
		for _, tag := range tags[m] {
			fmt.Println(tag)
		}
		return true
	})

	// Output:
	// websearch
	// google
	// subdomain
}
