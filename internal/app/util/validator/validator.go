package validator

import "regexp"

// Constants for validation.
const (
	URLRegex = "^(?:http|http(s)):\\/\\/(www\\.)?[-a-zA-Z0-9@:%._\\+~#=]{1,256}(\\.|:)[a-zA-Z0-9()]{1,6}\\b([-a-zA-Z0-9()@:%_\\+.~#?&//=]*)"

	IDRegex = "^[A-Za-z0-9]{8}$"
)

var urlMatcher = regexp.MustCompile(URLRegex)
var idMatcher = regexp.MustCompile(IDRegex)

// URL Validates URL.
func URL(url string) bool {
	return urlMatcher.MatchString(url)
}

// ID validates short ID.
func ID(url string) bool {
	return idMatcher.MatchString(url)
}
