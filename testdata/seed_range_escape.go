package testdata

func rangeEscape(vals []int) []*int {
	var ptrs []*int
	for _, v := range vals {
		ptrs = append(ptrs, &v)
	}
	return ptrs
}
