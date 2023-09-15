package validator

import "regexp"

const (
	URLRegex = "(?:http|http(s)):\\/\\/(www\\.)?[-a-zA-Z0-9@:%._\\+~#=]{1,256}(\\.|:)[a-zA-Z0-9()]{1,6}\\b([-a-zA-Z0-9()@:%_\\+.~#?&//=]*)"

	IDRegex = "^[A-Za-z]{8}"
)

var urlMatcher = regexp.MustCompile(URLRegex)
var idMatcher = regexp.MustCompile(IDRegex)

func URL(url string) bool {
	return urlMatcher.MatchString(url)
}

func ID(url string) bool {
	return idMatcher.MatchString(url)
}
