package fixer

import (
	"regexp"
	"strings"
)

var URLRegex = regexp.MustCompile(`https?:\/\/(?:www\.)?[-a-zA-Z0-9\._]+\/?[-a-zA-Z0-9()@:%_\+.~#?&\/=]+`)
var DomainRegex = regexp.MustCompile(`(?:https?:\/\/)?(?:www\.)?([-a-zA-Z0-9\._]+)\/?`)

func ExtractURLs(text string) []string {
	matches := URLRegex.FindAllStringSubmatch(text, -1)
	var urls []string
	for _, match := range matches {
		urls = append(urls, match[0])
	}
	return urls
}

func ExtractDomain(url string) string {
	match := DomainRegex.FindStringSubmatch(url)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func RemoveQueryParams(url string) string {
	//  Chop off query params
	url, _, _ = strings.Cut(url, "?")
	return url
}
