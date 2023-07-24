package core

// Filter - split slice in 2 - matching & not matching to a condition - TODO: Move to slice package
func Filter[T any](items []T, f func(t T) bool) (result []T, removed []T) {
	for _, item := range items {
		if f(item) {
			result = append(result, item)
		} else {
			removed = append(removed, item)
		}
	}
	return result, removed
}
