package lib

func ProcessVersion(version string) string {
	if version == "latest" {
		return ""
	}
	return version
}

func In(str string, list []string) bool {
	for _, s := range list {
		if s == str {
			return true
		}
	}
	return false
}

func Intersect(listA, listB []string) (list []string) {
	mapA := make(map[string]bool)
	for _, item := range listA {
		mapA[item] = true
	}
	for _, item := range listB {
		if mapA[item] {
			list = append(list, item)
		}
	}
	return list
}

func Unique(list []string) []string {
	was := make(map[string]bool)
	result := []string{}
	for _, v := range list {
		if !was[v] {
			was[v] = true
			result = append(result, v)
		}
	}
	return result
}

func Union(listA []string, listB []string) []string {
	listA = append(listA, listB[:]...)
	return Unique(listA)
}
