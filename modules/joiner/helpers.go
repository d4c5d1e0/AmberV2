package joiner

func getRanges(index, multiplier, memberCount int) [][]int { // https://github.com/Merubokkusu/Discord-S.C.U.M/blob/77daf74354415cb5d9411f886899c9817d0bc5b9/discum/gateway/guild/combo.py#L48
	initalNum := index * multiplier
	rangesList := [][]int{{initalNum, initalNum + 99}}

	if memberCount > initalNum+99 {
		rangesList = append(rangesList, []int{initalNum + 100, initalNum + 199})
	}
	if !belongToIntSlice(rangesList, []int{0, 99}) {
		rangesList = append(rangesList, []int{})
		insert(rangesList, []int{0, 99}, 0)
	}
	return rangesList
}

func belongsToStrSlice(input []string, lookup string) bool { // https://stackoverflow.com/a/52710077
	for _, val := range input {
		if val == lookup {
			return true
		}
	}
	return false
}

func belongToIntSlice(input [][]int, lookup []int) bool { // https://stackoverflow.com/a/52710077
	for _, val := range input {
		if Equal(val, lookup) {
			return true
		}
	}
	return false
}

func insert(a [][]int, c []int, i int) [][]int { //https://github.com/golang/go/wiki/SliceTricks#insert
	return append(a[:i], append([][]int{c}, a[i:]...)...)
}

func Equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
