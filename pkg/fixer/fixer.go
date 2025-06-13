package fixer

import (
	"encoding/gob"
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/log"
)

// init registers the concrete Fixer types with gob so they can be serialized.
func init() {
	gob.Register(ReplaceFixer{})
	gob.Register(RegexpReplaceFixer{})
	gob.Register(PrependFixer{})
}

// A Fixer fixes an input URL and returns a corrected copy.
type Fixer interface {
	String() string
	Fix(string) string
}

// A ReplaceFixer performs simple replacement on its URL.
type ReplaceFixer struct {
	Old string
	New string
}

// Fix replaces all instances of f.Old in link with f.New.
func (f ReplaceFixer) Fix(link string) string {
	return strings.ReplaceAll(link, f.Old, f.New)
}

func (f ReplaceFixer) String() string {
	return fmt.Sprintf("replace '%v' with '%v'", f.Old, f.New)
}

// A RegexpReplaceFixer replaces matches of a regular expression
// with a replacement string.
type RegexpReplaceFixer struct {
	Pattern     string
	Replacement string
}

// Fix replaces all matching instances of f.Pattern in link
// with f.Replacement.
//
// Capture groups can accessed in f.Replacement using $x
// following regexp.Regexp.Expand syntax.
func (f RegexpReplaceFixer) Fix(link string) string {
	re, err := regexp.Compile(f.Pattern)
	if err != nil {
		log.Error("could not compile regex", "err", err)
		return ""
	}
	return re.ReplaceAllString(link, f.Replacement)
}

func (f RegexpReplaceFixer) String() string {
	return fmt.Sprintf("regex replace '%v' with '%v'", f.Pattern, f.Replacement)
}

type PrependFixer struct {
	Prefix string
}

func (f PrependFixer) String() string {
	return fmt.Sprintf("prepend '%v'", f.Prefix)
}

func (f PrependFixer) Fix(link string) string {
	return f.Prefix + link
}
