package trie

const defaultMask = '*'

type (
	Trie interface {
		Filter(text string) (maskText string, keywords []string, hit bool)
		Keywords(text string) (hits []string)
	}

	trieNode struct {
		Node
		mask rune
	}

	scope struct {
		indexFrom int
		indexTo   int
	}

	Option func(node *trieNode)
)

// WithMask set trie tree mask
func WithMask(mask rune) Option {
	return func(node *trieNode) {
		node.mask = mask
	}
}

// New return a trie tree with keywords
func New(keywords []string, opts ...Option) Trie {
	t := new(trieNode)

	for _, opt := range opts {
		opt(t)
	}

	if t.mask == 0 {
		t.mask = defaultMask
	}

	for _, keyword := range keywords {
		t.Node.Add(keyword)
	}

	return t
}

// Filter would find the keyword from original text, and replace keyword with mask
func (t *trieNode) Filter(text string) (maskText string, keywords []string, hit bool) {
	chars := []rune(text)
	if len(chars) == 0 {
		return text, nil, false
	}

	scopes := t.keywordScopes(chars)
	keywords = t.keywords(chars, scopes)

	for _, scope := range scopes {
		t.replaceWithMask(chars, scope.indexFrom, scope.indexTo)
	}

	return string(chars), keywords, len(keywords) > 0
}

// Keywords return the keywords which hit from original trie tree
func (t *trieNode) Keywords(text string) (hits []string) {
	chars := []rune(text)
	if len(chars) == 0 {
		return nil
	}

	return t.keywords(chars, t.keywordScopes(chars))
}

func (t *trieNode) keywords(chars []rune, scopes []scope) []string {
	set := make(map[string]struct{})
	for _, v := range scopes {
		set[string(chars[v.indexFrom:v.indexTo])] = struct{}{}
	}

	hits := make([]string, 0, len(set))
	for k := range set {
		hits = append(hits, k)
	}

	return hits
}

func (t *trieNode) keywordScopes(chars []rune) (scopes []scope) {
	var (
		size      = len(chars)
		indexFrom = -1
	)

	for i := 0; i < size; i++ {
		// find every single char from root node
		child, ok := t.Node.Children[chars[i]]
		if !ok {
			continue
		}

		if indexFrom < 0 {
			indexFrom = i
		}

		// last node
		if child.IsEnd() {
			scopes = append(scopes, scope{
				indexFrom: indexFrom,
				indexTo:   i + 1,
			})
		}

		// find the longest string that matches
		for j := i + 1; j < size; j++ {
			grandchild, ok := child.Children[chars[j]]
			if !ok {
				break
			}

			child = grandchild
			if child.IsEnd() {
				scopes = append(scopes, scope{
					indexFrom: indexFrom,
					indexTo:   j + 1,
				})
			}
		}

		indexFrom = -1
	}

	return scopes
}

func (t *trieNode) replaceWithMask(chars []rune, from, to int) {
	for i := from; i < to; i++ {
		chars[i] = t.mask
	}
}
