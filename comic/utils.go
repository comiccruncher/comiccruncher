package comic

// must panics if the err is not nil.
func must(err error) {
	if err != nil {
		panic(err)
	}
}
