package http

import (
	"io"
	"net/http"
	"strings"

	"github.com/sado0823/go-kitx/transport/http/read"
)

const (
	baseContentType = "application"
)

func contentType(subtype string) string {
	return strings.Join([]string{baseContentType, subtype}, "/")
}

// according rfc7231. form-data, x-www-form-urlencoded,json
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

func rawData(contentType string) rawFunc {
	contentSubtype := contentSubtype(contentType)
	fn, ok := raws[contentSubtype]
	if ok {
		return fn
	}
	return raws["json"]
}

type rawFunc func(*http.Request) ([]byte, error)

var (
	rawForm = func(r *http.Request) ([]byte, error) {
		if r.Method == http.MethodGet {
			if r.URL.RawQuery != "" {
				// try read from query.
				return []byte(read.GetQuery(r).Encode()), nil
			}
		}
		// otherwise use the ReadForm,
		// it's actually the same except
		// ReadQuery will not fire errors on:
		// 1. unknown or empty url query parameters
		// 2. empty query or form (if FireEmptyFormError is enabled).
		// 32M
		form, _ := read.GetForm(r, 32<<20, false)
		return []byte(form.Encode()), nil
	}

	// map[content_sub_type]func
	raws = map[string]rawFunc{
		"json": func(r *http.Request) ([]byte, error) {
			return io.ReadAll(r.Body)
		},
		"x-www-form-urlencoded": rawForm,
		"form-data":             rawForm,
	}
)
