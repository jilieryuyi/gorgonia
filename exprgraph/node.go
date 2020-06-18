package exprgraph

import (
	"fmt"
	"log"

	"gorgonia.org/gorgonia"
	"gorgonia.org/gorgonia/execution"
	"gorgonia.org/gorgonia/ops"
	"gorgonia.org/tensor"
)

// constraints
var (
	_ gorgonia.Tensor = &Node{}
)

// NodeID represents a Node's ID
type NodeID int64

// ID returns the ID as an int64. This is used to fulfil gonum.org/gonum/graph.Node interface.
func (n NodeID) ID() int64 { return int64(n) }

// Node is a tuple of a Tensor, ID, and name.
type Node struct {
	gorgonia.Tensor
	id   int64
	name string
}

func Make(g *Graph, name string, opts ...tensor.ConsOpt) Node {
	var eng tensor.Engine = g.StandardEngine
	if _, ok := g.StandardEngine.(tensor.StdEng); ok {
		eng = g
	}
	consOpts := append([]tensor.ConsOpt{tensor.WithEngine(eng), inGraph(), WithName(name)}, opts...)
	t := tensor.New(consOpts...)
	return Cons(g, name, t)
}

func Cons(g *Graph, name string, t tensor.Tensor) Node {
	id := g.idOrInsert(t)
	g.nodes[id].name = name
	return g.nodes[id].Node
}

// OK returns true if the Node is good for processing.
func (n *Node) OK() bool { return n.Tensor != nil }

// ID allows Node  to implement gonum.org/graph.Node
func (n *Node) ID() int64 { return n.id }

// Node implements gorgonia.Result

func (n *Node) Node() Node   { return *n }
func (n *Node) Nodes() Nodes { return Nodes{*n} }
func (n *Node) Err() error   { return nil }

func (n Node) Format(f fmt.State, c rune) {
	switch c {
	case 's':
		fmt.Fprintf(f, "%s", n.name)
	default:
		switch t := n.Tensor.(type) {
		case tensor.Tensor:
			str := consFmtStr(f, c)
			fmt.Fprintf(f, str, t)
		default:
			log.Printf("tensor type %T unsupported for node.Format", n.Tensor)
		}
	}
}

// GraphNode is a tuple of a graph object and a node. This allows for querying the payload of the Node.
//
// This is the object that should be used for any kind of query (topsort, etc)
type GraphNode struct {
	*Graph
	Node
}

//go:notinheap
type gn struct {
	*Graph
	Node
}

// node is a node for internal use. Its graph is defined by the links (i.e. pointers).
// if the ID is negative, it means that the node is in-progress
type node struct {
	Node
	flag
	execution.Device
	ops.Op
}
