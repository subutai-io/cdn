package app

func FileType(scope int) string {
	if scope == PublicScope {
		return "public"
	} else if scope == PrivateScope {
		return "private"
	}
	return ""
}

func ScopeType(scope int) string {
	if scope == PublicScope {
		return "false"
	} else if scope == PrivateScope {
		return "true"
	}
	return ""
}

func SlicesEqual(listA []*OldResult, listB []*OldResult) bool {
	for _, fileA := range listA {
		exists := false
		for i, fileB := range listB {
			if fileA == fileB {
				exists = true
				listB = append(listB[:i], listB[i + 1:]...)
				break
			}
		}
		if !exists {
			return false
		}
	}
	return true
}
