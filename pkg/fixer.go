package linkfixerbot

import (
	"encoding/gob"
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/log"
)

func init() {
	gob.Register(ReplaceFixer{})
	gob.Register(ReplaceRegexFixer{})
	gob.Register(PrependFixer{})
}

type Fixer interface {
	String() string
	Fix(string) string
}

type ReplaceFixer struct {
	Old string
	New string
}

func (f ReplaceFixer) Fix(link string) string {
	return strings.ReplaceAll(link, f.Old, f.New)
}

func (f ReplaceFixer) String() string {
	return fmt.Sprintf("replace '%v' with '%v'", f.Old, f.New)
}

type ReplaceRegexFixer struct {
	Pattern     string
	Replacement string
}

func (f ReplaceRegexFixer) Fix(link string) string {
	re, err := regexp.Compile(f.Pattern)
	if err != nil {
		log.Error("could not compile regex", "err", err)
		return ""
	}
	return re.ReplaceAllString(link, f.Replacement)
}

func (f ReplaceRegexFixer) String() string {
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
