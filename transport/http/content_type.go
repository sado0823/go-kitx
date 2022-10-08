package http

import "strings"

const (
	baseContentType = "application"
)

func contentType(subtype string) string {
	return strings.Join([]string{baseContentType, subtype}, "/")
}

// according rfc7231.
func contentSubtype(contentType string) string {
	left := strings.Index(contentType, "/")
	if left == -1 {
		return ""
	}
	right := strings.Index(contentType, ";")
	if right == -1 {
		right = len(contentType)
	}
	if right < left {
		return ""
	}
	return contentType[left+1 : right]
}
