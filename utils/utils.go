package utils

func ProcessVersion(version string) string {
	if version == "latest" {
		return ""
	}
	return version
}

func In(str []string, list []string) bool {
	for _, s := range list {
		for _, t := range str {
			if s == t {
				return true
			}
		}
	}
	return false
}

func Intersect(listA, listB []string) (list []string) {
	mapA := make(map[string]int)
	for _, item := range listA {
		mapA[item]++
	}
	for _, item := range listB {
		if mapA[item] > 0 {
			mapA[item]--
			list = append(list, item)
		}
	}
	return
}

func Unique(list []string) (result []string) {
	was := make(map[string]bool)
	for _, v := range list {
		if !was[v] {
			was[v] = true
			result = append(result, v)
		}
	}
	return
}

func Union(listA []string, listB []string) []string {
	listA = append(listA, listB[:]...)
	return Unique(listA)
}
