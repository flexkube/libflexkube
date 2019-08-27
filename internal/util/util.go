package util

// DefaultString returns value if it is not empty
//
// If value is empty, default value is returned
func DefaultString(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}

	return value
}
