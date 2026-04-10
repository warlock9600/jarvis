package common

import "strings"

func ShouldMaskKey(key string) bool {
	k := strings.ToLower(key)
	return strings.Contains(k, "password") ||
		strings.Contains(k, "token") ||
		strings.Contains(k, "secret") ||
		strings.Contains(k, "apikey") ||
		strings.Contains(k, "api_key")
}

func MaskValue(value string) string {
	if value == "" {
		return ""
	}
	return "********"
}
