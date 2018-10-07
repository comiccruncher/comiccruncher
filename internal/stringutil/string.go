package stringutil

import (
	"math/rand"
	"strings"
	"sync"
	"time"
)

const randCharMap = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890"

var once sync.Once

// Generate the random seed at program start.
func init() {
	once.Do(func() {
		rand.Seed(time.Now().UnixNano())
	})
}

// Checks if any of the given `prefixes` are in the `s` string.
func HasAnyPrefix(s string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

// Checks if any of the given `prefixes` are in the `s` string.
// Case insensitive.
func HasAnyiPrefix(s string, prefixes ...string) bool {
	s = strings.ToLower(s)
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, strings.ToLower(prefix)) {
			return true
		}
	}
	return false
}

// Checks if the string `s` is equal to any of the strings `strs`.
// Case insensitive and trims the strings.
func EqualsIAny(s string, strs ...string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	for _, str := range strs {
		if s == strings.TrimSpace(strings.ToLower(str)) {
			return true
		}
	}
	return false
}

// Returns true if the string is empty or nil.
func Empty(s *string) bool {
	if s == nil {
		return true
	}
	if *s == "" {
		return true
	}
	return false
}

// Generates a random string of the length `n` with characters A-Z, a-z, and 0-9.
func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = randCharMap[rand.Intn(len(randCharMap))]
	}
	rand.Int63()
	return string(b)
}
