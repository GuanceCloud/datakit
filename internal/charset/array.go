package charset

func Contains(set []string, elem string) bool {
	for _, v := range set {
		if v == elem {
			return true
		}
	}

	return false
}

func Differ(source, compare []string) []string {
	m := make(map[string]struct{}, len(compare))
	for _, v := range compare {
		m[v] = struct{}{}
	}

	var diff []string
	for _, v := range source {
		if _, found := m[v]; !found {
			diff = append(diff, v)
		}
	}

	return diff
}
