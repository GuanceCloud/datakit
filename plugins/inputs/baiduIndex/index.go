package baiduIndex

func splitWord(words []string) [][]string {
	var list [][]string
	var item = []string{}
	for idx, word := range words {
		if (idx+1)%5 == 0 {
			list = append(list, item)
		} else {
			item = append(item, word)
		}
	}

	return list
}
