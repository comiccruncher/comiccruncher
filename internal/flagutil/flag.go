package flagutil

import (
	"github.com/spf13/pflag"
	"strings"
)

// Split splits a flag and trims whitespace by the specified separator `sep`.
func Split(flag pflag.Flag, sep string) []string {
	var results []string
	if flag.Value.String() != "" {
		results = strings.Split(flag.Value.String(), sep)
		for i, p := range results {
			results[i] = strings.TrimSpace(p)
		}
	}
	return results
}
