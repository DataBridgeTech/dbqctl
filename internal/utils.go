package internal

func IdOrDefault(original string, defaultVal string) string {
	if original == "" {
		return defaultVal
	}
	return original
}
