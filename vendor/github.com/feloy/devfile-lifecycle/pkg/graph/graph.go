package graph

import (
	"sort"

	"github.com/Heiko-san/mermaidgen/flowchart"
)

type Graph struct {
	EntryNodeID string
	nodes       map[string]*Node
	edges       []*Edge
}

func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[string]*Node),
	}
}

func (o *Graph) AddNode(id string, text ...string) *Node {
	node := Node{
		ID:   id,
		Text: text,
	}
	o.nodes[id] = &node
	return &node
}

func (o *Graph) AddEdge(from *Node, to *Node, text ...string) *Edge {
	edge := Edge{
		From: from,
		To:   to,
		Text: text,
	}
	o.edges = append(o.edges, &edge)
	return &edge
}

func (o *Graph) ToFlowchart() *flowchart.Flowchart {
	f := flowchart.NewFlowchart()
	n := f.AddNode(o.EntryNodeID)
	n.Text = o.nodes[o.EntryNodeID].Text
	keys := make([]string, 0, len(o.nodes))
	for k := range o.nodes {
		keys = append(keys, k)

	}
	sort.Strings(keys)
	for _, key := range keys {
		node := o.nodes[key]
		if node.ID == o.EntryNodeID {
			continue
		}
		n := f.AddNode(node.ID)
		n.Text = node.Text
	}

	for _, edge := range o.edges {
		e := f.AddEdge(f.GetNode(edge.From.ID), f.GetNode(edge.To.ID))
		e.Text = edge.Text
	}
	return f
}

type Node struct {
	ID   string
	Text []string
}

type Edge struct {
	From *Node
	To   *Node
	Text []string
}
