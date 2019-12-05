package util

// PickString returns first non-empty string passed.
func PickString(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}

	return ""
}

// PickInt returns first non-zero integer passed.
func PickInt(values ...int) int {
	for _, v := range values {
		if v != 0 {
			return v
		}
	}

	return 0
}
