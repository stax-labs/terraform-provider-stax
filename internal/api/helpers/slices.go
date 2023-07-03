package helpers

//	Subtract takes two maps and returns a slice of keys that exist in the left map but not the right.
//
// left is the map to subtract from.
//
// right is the map to subtract.
//
// The function returns a slice of keys that exist in left but not right.
func Subtract(left map[string]bool, right map[string]bool) []string {
	remainder := make(map[string]bool)

	for key := range left {
		if _, ok := right[key]; !ok {
			remainder[key] = true
		}
	}

	s := []string{}

	for k := range remainder {
		s = append(s, k)
	}

	return s
}
