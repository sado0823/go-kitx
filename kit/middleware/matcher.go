package middleware

import (
	"sort"
	"strings"
)

type Matcher interface {
	Use(ms ...Middleware)
	Add(selector string, ms ...Middleware)
	Match(operation string) []Middleware
}

func NewMatcher() Matcher {
	return &matcher{
		matches: make(map[string][]Middleware),
	}
}

type matcher struct {
	prefix   []string
	defaults []Middleware
	matches  map[string][]Middleware
}

func (m *matcher) Use(ms ...Middleware) {
	m.defaults = ms
}

func (m *matcher) Add(selector string, ms ...Middleware) {
	if strings.HasSuffix(selector, "*") {
		selector = strings.TrimSuffix(selector, "*")
		m.prefix = append(m.prefix, selector)
		// sort the prefix:
		//  - /foo/bar
		//  - /foo
		sort.Slice(m.prefix, func(i, j int) bool {
			return m.prefix[i] > m.prefix[j]
		})
	}

	m.matches[selector] = ms
}

func (m *matcher) Match(operation string) []Middleware {
	ms := make([]Middleware, 0, len(m.defaults))
	ms = append(ms, m.defaults...)

	if next, ok := m.matches[operation]; ok {
		return append(ms, next...)
	}

	for _, prefix := range m.prefix {
		if strings.HasPrefix(operation, prefix) {
			return append(ms, m.matches[prefix]...)
		}
	}

	return ms
}
