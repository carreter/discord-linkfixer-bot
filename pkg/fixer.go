package linkfixerbot

import (
	"encoding/gob"
	"regexp"
	"strings"

	"github.com/charmbracelet/log"
)

func init() {
	gob.Register(ReplaceFixer{})
	gob.Register(ReplaceRegexFixer{})
	gob.Register(PrependerFixer{})
}

type Fixer interface {
	Fix(string) string
}

type ReplaceFixer struct {
	Old string
	New string
}

func (f ReplaceFixer) Fix(link string) string {
	return strings.ReplaceAll(link, f.Old, f.New)
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

type PrependerFixer struct {
	Prefix string
}

func (f PrependerFixer) Fix(link string) string {
	return f.Prefix + link
}
