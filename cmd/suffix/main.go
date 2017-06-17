package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/tg/suffix"
)

func main() {
	var (
		fromfile   string
		fixedNames bool
		fieldDelim string
		fieldNum   int

		filter Filter
	)

	// Main application
	app := &cobra.Command{
		Use: "suffix pattern [file...]",
		Long: `suffix -- domain name suffix search and print

Suffix tool scans input lines and tries to locate a column containing name matching
suffix pattern. Suffix matches the name if the latter ends with the same labels
as specified by suffix. Labels are delimited by dots, as in DNS.

For example suffix 'golang.org' would match 'blog.golang.org', but wouldn't
match 'amigolang.org'.

If no input files are provided, lines are read from stdin.
Multiple name patterns can be provided by using -f flag, in which case
no patterns are expected on command line.

All matching is case sensitive.
`,
		Run: func(cmd *cobra.Command, args []string) {
			log.SetFlags(0)
			var sfx suffix.Set

			if fromfile != "" {
				f, err := os.Open(fromfile)
				if err != nil {
					log.Fatalf("%s: %s", fromfile, err)
				}
				_, err = sfx.ReadFrom(f)
				f.Close()
			} else if len(args) > 0 {
				// Take name filter from arguments
				sfx.Add(args[0])
				// Remove name from arguments
				args = args[1:]
			} else {
				cmd.Help()
				os.Exit(2)
			}

			// Prepare data inputs -- either from files or stdin
			var inputs []io.ReadCloser
			for _, path := range args {
				f, err := os.Open(path)
				if err != nil {
					log.Fatalf("%s: %s", path, err)
				}
				inputs = append(inputs, f)
			}
			if len(inputs) == 0 {
				inputs = append(inputs, ioutil.NopCloser(os.Stdin))
			}

			if fixedNames {
				filter.MatchField = func(name string) string {
					if sfx.MatchesExact(name) {
						return name
					}
					return ""
				}
			} else {
				filter.MatchField = sfx.Match
			}

			for _, in := range inputs {
				var scan LineScanner
				scan = NewSplitScanner(in, fieldDelim)
				if fieldNum > 0 {
					scan = &SingleFieldScanner{scan, fieldNum - 1}
				}
				err := filter.Run(scan, os.Stdout)
				if err != nil {
					log.Fatal(err)
				}
				in.Close()
			}
		},
	}
	app.Flags().StringVarP(&fromfile, "file", "f", "", "read name patterns from file")
	app.Flags().BoolVarP(&filter.MatchNone, "invert-match", "v", false, "invert match")
	app.Flags().BoolVarP(&fixedNames, "fixed-names", "F", false, "treat suffixes as FQDNs, so they must match exactly")
	app.Flags().StringVarP(&fieldDelim, "delimiter", "d", "\t", "column delimiter")
	app.Flags().IntVarP(&fieldNum, "column", "c", 0, "select only one column for matching")
	app.Flags().BoolVar(&filter.OnlyMatch, "only-suffix", false, "print only matching suffix")
	app.Flags().BoolVar(&filter.OnlyField, "only-name", false, "print only matching name")

	app.Execute()
}
