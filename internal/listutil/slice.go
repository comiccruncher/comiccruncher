package listutil

// StringKeys returns the keys for a map string.
func StringKeys(m map[string]string) []string {
	keys := make([]string, 0)
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// StringInSlice checks if the string `s` is in the slice `strs`.
func StringInSlice(strs []string, s string) bool {
	for _, st := range strs {
		if st == s {
			return true
		}
	}
	return false
}
