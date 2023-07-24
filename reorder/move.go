package reorder

// MoveStr moves a string in an array
func MoveStr(arr []string, from, to int) []string {
	if from == to {
		return arr
	}
	a := make([]string, len(arr))
	copy(a, arr)
	item := a[from]
	if from < to {
		for i := from; i < to; i++ {
			a[i] = a[i+1]
		}
	} else if to < from {
		for i := from; i > to; i-- {
			a[i] = a[i-1]
		}
	}
	a[to] = item
	return a
}

// RemoveStr removes a string from an array
func RemoveStr(arr []string, s string) []string {
	a := make([]string, 0, len(arr))
	for _, v := range arr {
		if v != s {
			a = append(a, v)
		}
	}
	return a
}

// InsertStr insert a string into an array
func InsertStr(arr []string, index int, v string) []string {
	a := make([]string, len(arr)+1)
	a = append(a, arr[:index]...)
	a = append(a, v)
	return append(a, arr[index:]...)
}

// IndexOf returns index of a string in an array or -1 if not found
func IndexOf(arr []string, s string) int {
	for i, v := range arr {
		if v == s {
			return i
		}
	}
	return -1
}
