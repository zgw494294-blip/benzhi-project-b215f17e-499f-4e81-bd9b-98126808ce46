package transport

import (
	"errors"
	"net/url"
	"strings"
)

func pathParts(path string) []string {
	p := strings.Trim(path, "/")
	if p == "" {
		return nil
	}
	return strings.Split(p, "/")
}
func queryValue(u *url.URL, key string) string { return u.Query().Get(key) }

var errMalformedPath = errors.New("malformed path")
