package main

import (
	"log"
	"strconv"
	"strings"
)

// parseURL parses host as an authority without user
// information. That is, as host[:port].
func parseURL(host string, port string) (string, string, error) {
	if i := strings.LastIndex(host, ":"); i != -1 {
		colonPort := host[i:]
		if strings.HasPrefix(colonPort, "]") {
			// no port present
			return host, port, nil
		}
		if !validOptionalPort(colonPort) {
			log.Fatalf("invalid port %q after host", colonPort[1:])
		}
		return host[:i], colonPort[1:], nil
	} else {
		return host, port, nil
	}
}

// validOptionalPort reports whether port is either an empty string
// or matches /^:\d*$/
func validOptionalPort(port string) bool {
	if port == "" {
		return true
	}
	if port[0] != ':' {
		return false
	}
	for _, b := range port[1:] {
		if b < '0' || b > '9' {
			return false
		}
	}
	if a, _ := strconv.Atoi(port[1:]); a >= 65536 || a < 0 {
		return false
	}
	return true
}

// // unescape unescapes a string; the mode specifies
// // which section of the URL string is being unescaped.
// func unescape(s string, mode encoding) (string, error) {
// 	// Count %, check that they're well-formed.
// 	n := 0
// 	hasPlus := false
// 	for i := 0; i < len(s); {
// 		switch s[i] {
// 		case '%':
// 			n++
// 			if i+2 >= len(s) || !ishex(s[i+1]) || !ishex(s[i+2]) {
// 				s = s[i:]
// 				if len(s) > 3 {
// 					s = s[:3]
// 				}
// 				return "", EscapeError(s)
// 			}
// 			// Per https://tools.ietf.org/html/rfc3986#page-21
// 			// in the host component %-encoding can only be used
// 			// for non-ASCII bytes.
// 			// But https://tools.ietf.org/html/rfc6874#section-2
// 			// introduces %25 being allowed to escape a percent sign
// 			// in IPv6 scoped-address literals. Yay.
// 			if mode == encodeHost && unhex(s[i+1]) < 8 && s[i:i+3] != "%25" {
// 				return "", EscapeError(s[i : i+3])
// 			}
// 			if mode == encodeZone {
// 				// RFC 6874 says basically "anything goes" for zone identifiers
// 				// and that even non-ASCII can be redundantly escaped,
// 				// but it seems prudent to restrict %-escaped bytes here to those
// 				// that are valid host name bytes in their unescaped form.
// 				// That is, you can use escaping in the zone identifier but not
// 				// to introduce bytes you couldn't just write directly.
// 				// But Windows puts spaces here! Yay.
// 				v := unhex(s[i+1])<<4 | unhex(s[i+2])
// 				if s[i:i+3] != "%25" && v != ' ' && shouldEscape(v, encodeHost) {
// 					return "", EscapeError(s[i : i+3])
// 				}
// 			}
// 			i += 3
// 		case '+':
// 			hasPlus = mode == encodeQueryComponent
// 			i++
// 		default:
// 			if (mode == encodeHost || mode == encodeZone) && s[i] < 0x80 && shouldEscape(s[i], mode) {
// 				return "", InvalidHostError(s[i : i+1])
// 			}
// 			i++
// 		}
// 	}

// 	if n == 0 && !hasPlus {
// 		return s, nil
// 	}

// 	var t strings.Builder
// 	t.Grow(len(s) - 2*n)
// 	for i := 0; i < len(s); i++ {
// 		switch s[i] {
// 		case '%':
// 			t.WriteByte(unhex(s[i+1])<<4 | unhex(s[i+2]))
// 			i += 2
// 		case '+':
// 			if mode == encodeQueryComponent {
// 				t.WriteByte(' ')
// 			} else {
// 				t.WriteByte('+')
// 			}
// 		default:
// 			t.WriteByte(s[i])
// 		}
// 	}
// 	return t.String(), nil
// }
