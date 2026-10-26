package checker

import (
	"net/url"
	"regexp"
	"strings"
)

// sameURL reports whether two URLs name the same resource under the
// normalizations of RFC 3986, section 6.2.2 and 6.2.3: the scheme, the host
// and percent-encoding hex digits are case-insensitive, a percent-encoded
// unreserved character equals the character, and an http or https URL may
// omit its default port and an empty path. The path, query and fragment are
// otherwise compared as written.
func sameURL(from, to any) bool {
	fromURL, fromOk := from.(string)
	toURL, toOk := to.(string)
	if !fromOk || !toOk {
		return false
	}
	return normalizeURL(fromURL) == normalizeURL(toURL)
}

var percentEncoding = regexp.MustCompile(`%[0-9a-fA-F]{2}`)

func normalizePercentEncoding(s string) string {
	return percentEncoding.ReplaceAllStringFunc(s, func(encoded string) string {
		decoded, err := url.PathUnescape(encoded)
		if err == nil && isUnreserved(decoded[0]) {
			return decoded
		}
		return strings.ToUpper(encoded)
	})
}

func isUnreserved(c byte) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' || c == '-' || c == '.' || c == '_' || c == '~'
}

func normalizeURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		// A relative reference has no scheme or host to normalize.
		return normalizePercentEncoding(rawURL)
	}

	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	if port := u.Port(); u.Scheme == "https" && port == "443" || u.Scheme == "http" && port == "80" {
		u.Host = strings.TrimSuffix(u.Host, ":"+port)
	}
	if u.Path == "" && (u.Scheme == "http" || u.Scheme == "https") {
		u.Path = "/"
	}

	return normalizePercentEncoding(u.String())
}
