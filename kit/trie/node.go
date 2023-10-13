package trie

// Node is a trie tree node
type Node struct {
	Children map[rune]*Node
	end      bool
}

// IsEnd check this node is the last node (with no children)
func (n *Node) IsEnd() bool {
	return n.end
}

// Add add incoming string to trie tree
func (n *Node) Add(word string) {
	chars := []rune(word)
	if len(chars) == 0 {
		return
	}

	// nd is the last node
	nd := n
	for _, char := range chars {
		if nd.Children == nil {
			// have no children
			nd.Children = make(map[rune]*Node)
			newN := new(Node)
			nd.Children[char] = newN
			// next char would follow this node
			nd = newN

		} else if v, ok := nd.Children[char]; ok {
			// this char is already in trie tree
			nd = v

		} else {
			// have children, but not cotain this char
			newN := new(Node)
			nd.Children[char] = newN
			// next char would follow this node
			nd = newN
		}
	}

	nd.end = true

}
