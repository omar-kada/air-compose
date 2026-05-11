package testutil

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

// CookiesToMap converts a slice of http.Cookie pointers to a map with cookie names as keys.
func CookiesToMap(cookies []*http.Cookie) map[string]*http.Cookie {
	cookiesMap := make(map[string]*http.Cookie)
	for _, cookie := range cookies {
		cookiesMap[cookie.Name] = cookie
	}
	return cookiesMap
}

// AssertCookiesAreSecure verifies that specified cookies in the response are marked as secure.
func AssertCookiesAreSecure(t *testing.T, rr *httptest.ResponseRecorder, cookieNames ...string) {
	cookies := rr.Result().Cookies()

	for _, cookie := range cookies {
		if slices.Contains(cookieNames, cookie.Name) {
			assert.True(t, cookie.Secure, cookie.Name+" cookie should be secure")
		}
	}
}

// AssertCookiesAreNotSecure verifies that specified cookies in the response are not marked as secure.
func AssertCookiesAreNotSecure(t *testing.T, rr *httptest.ResponseRecorder, cookieNames ...string) {
	cookies := rr.Result().Cookies()

	for _, cookie := range cookies {
		if slices.Contains(cookieNames, cookie.Name) {
			assert.False(t, cookie.Secure, cookie.Name+" cookie should not be secure")
		}
	}
}
