package flow

import (
	"bytes"
	"fmt"
	"log"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

func NewGraphviz() {
	g := graphviz.New()
	graph, err := g.Graph(graphviz.Name("test name"))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := graph.Close(); err != nil {
			log.Fatal(err)
		}
		g.Close()
	}()

	type todo struct {
		name  string
		label string
	}

	var (
		gateways = []todo{
			{name: "gateway_start", label: "开始审核"},
			{name: "gateway_end", label: "结束审核"},
		}
		circles = []todo{
			{name: "start", label: "开始"},
			{name: "end", label: "结束"},
		}
		boxs = []todo{
			{name: "company", label: "公司审核"},
			{name: "department", label: "部门审核"},
			{name: "user", label: "用户审核"},
		}
	)

	gatewayStart, err := Diamond(graph, gateways[0].name, gateways[0].label)
	if err != nil {
		log.Fatal(err)
	}
	gatewayEnd, err := Diamond(graph, gateways[1].name, gateways[1].label)
	if err != nil {
		log.Fatal(err)
	}

	start, err := Circle(graph, circles[0].name, circles[0].label)
	if err != nil {
		log.Fatal(err)
	}
	end, err := Circle(graph, circles[1].name, circles[1].label)
	if err != nil {
		log.Fatal(err)
	}

	company, err := Box(graph, boxs[0].name, boxs[0].label)
	if err != nil {
		log.Fatal(err)
	}
	department, err := Box(graph, boxs[1].name, boxs[1].label)
	if err != nil {
		log.Fatal(err)
	}
	user, err := Box(graph, boxs[2].name, boxs[2].label)
	if err != nil {
		log.Fatal(err)
	}

	err = Link(graph, "to_gateway_start", "开始审核", start, gatewayStart)
	if err != nil {
		log.Fatal(err)
	}

	err = Link(graph, "gateway_to_company", "公司条件审核流程", gatewayStart, company)
	if err != nil {
		log.Fatal(err)
	}

	err = Link(graph, "gateway_to_department", "部门条件审核流程", gatewayStart, department)
	if err != nil {
		log.Fatal(err)
	}

	err = Link(graph, "gateway_to_user", "用户条件审核流程", gatewayStart, user)
	if err != nil {
		log.Fatal(err)
	}

	err = Link(graph, "company_to_end", "公司审核流程", company, gatewayEnd)
	if err != nil {
		log.Fatal(err)
	}

	err = Link(graph, "department_to_end", "部门审核流程", department, gatewayEnd)
	if err != nil {
		log.Fatal(err)
	}

	err = Link(graph, "user_to_end", "用户审核流程", user, gatewayEnd)
	if err != nil {
		log.Fatal(err)
	}

	err = Link(graph, "to_end", "结束审核", gatewayEnd, end)
	if err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer
	if err := g.Render(graph, graphviz.XDOT, &buf); err != nil {
		log.Fatal(err)
	}
	fmt.Println(buf.String())

}

func Diamond(graph *cgraph.Graph, name string, label string) (*cgraph.Node, error) {
	return doNode(graph, name, label, cgraph.DiamondShape)
}

func Circle(graph *cgraph.Graph, name string, label string) (*cgraph.Node, error) {
	return doNode(graph, name, label, cgraph.CircleShape)
}

func Box(graph *cgraph.Graph, name string, label string) (*cgraph.Node, error) {
	return doNode(graph, name, label, cgraph.BoxShape)
}

func doNode(graph *cgraph.Graph, name string, label string, shape cgraph.Shape) (*cgraph.Node, error) {
	node, err := graph.CreateNode(name)
	if err != nil {
		return nil, err
	}
	node.SetLabel(label)
	node.SetShape(shape)

	return node, nil
}

func Link(graph *cgraph.Graph, name string, label string, from *cgraph.Node, to *cgraph.Node) error {
	e, err := graph.CreateEdge(name, from, to)
	if err != nil {
		return err
	}
	e.SetLabel(label)
	return nil
}
