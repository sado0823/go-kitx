package flow

import (
	"testing"

	"github.com/goccy/go-graphviz"
)

func TestNewGraphviz(t *testing.T) {
	NewGraphviz()
}

func Test_Parse_Graphviz(t *testing.T) {
	graph, err := graphviz.ParseBytes([]byte(DotTpl))
	if err != nil {
		t.Errorf("ParseBytes err=%+v", err)
	}
	defer graph.Close()
	t.Log("graph: ",
		graph.Name(),
		graph.Get("label"),
		graph.Get("comment"),
	)
	t.Log("")

	node := graph.FirstNode()
	for node != nil {
		t.Log("next.Name()",
			node.Name(),
			node.Get("shape"),
			node.Get("label"),
			node.Get("comment"),
		)

		edge := graph.FirstEdge(node)
		for edge != nil {
			edgeNode := edge.Node()
			if edgeNode != nil {
				t.Log("edgeNode: ",
					edgeNode.Name(),
					edgeNode.Get("shape"),
					edgeNode.Get("label"),
					edgeNode.Get("comment"),
				)
			}
			t.Log("Edge: ",
				edge.Get("label"),
				edge.Get("comment"),
			)

			edge = graph.NextEdge(edge, edgeNode)
		}

		node = graph.NextNode(node)
		t.Log("")
	}
}
