package coordinates

import (
	"github.com/SamyRai/cityFinder/lib/city"
)

type Spatial interface {
	Bounds() *city.Rect
}

type RTree struct {
	Root       *RNode
	MaxEntries int
}

type RNode struct {
	Parent   *RNode
	Children []*RNode
	Leaf     bool
	Entries  []*Entry
	Rect     *city.Rect
}

type Entry struct {
	Child *RNode
	Rect  *city.Rect
	Obj   Spatial
}

func NewRTree(maxEntries int) *RTree {
	return &RTree{
		Root:       &RNode{Leaf: true},
		MaxEntries: maxEntries,
	}
}

func (t *RTree) Insert(obj Spatial) {
	leaf := t.chooseLeaf(t.Root, obj)
	leaf.Entries = append(leaf.Entries, &Entry{Obj: obj, Rect: obj.Bounds()})

	if len(leaf.Entries) > t.MaxEntries {
		t.splitNode(leaf)
	} else {
		t.adjustTree(leaf)
	}
}

func (t *RTree) chooseLeaf(n *RNode, obj Spatial) *RNode {
	if n.Leaf {
		return n
	}

	bestChild := n.Children[0]
	minEnlargement := bestChild.Rect.Enlargement(obj.Bounds())
	minArea := bestChild.Rect.Area()

	for i := 1; i < len(n.Children); i++ {
		child := n.Children[i]
		enlargement := child.Rect.Enlargement(obj.Bounds())
		if enlargement < minEnlargement {
			minEnlargement = enlargement
			minArea = child.Rect.Area()
			bestChild = child
		} else if enlargement == minEnlargement {
			area := child.Rect.Area()
			if area < minArea {
				minArea = area
				bestChild = child
			}
		}
	}

	return t.chooseLeaf(bestChild, obj)
}

func (t *RTree) splitNode(n *RNode) {
	// Quadratic split
	var bestSeed1, bestSeed2 int
	maxWaste := -1.0
	for i := 0; i < len(n.Entries); i++ {
		for j := i + 1; j < len(n.Entries); j++ {
			waste := n.Entries[i].Rect.Union(n.Entries[j].Rect).Area() - n.Entries[i].Rect.Area() - n.Entries[j].Rect.Area()
			if waste > maxWaste {
				maxWaste = waste
				bestSeed1 = i
				bestSeed2 = j
			}
		}
	}

	// Create new nodes
	node1 := &RNode{Parent: n.Parent, Leaf: n.Leaf}
	node2 := &RNode{Parent: n.Parent, Leaf: n.Leaf}
	node1.Entries = append(node1.Entries, n.Entries[bestSeed1])
	node2.Entries = append(node2.Entries, n.Entries[bestSeed2])
	node1.Rect = n.Entries[bestSeed1].Rect
	node2.Rect = n.Entries[bestSeed2].Rect

	// Distribute remaining entries
	remaining := append(n.Entries[:bestSeed1], n.Entries[bestSeed1+1:bestSeed2]...)
	remaining = append(remaining, n.Entries[bestSeed2+1:]...)

	for _, entry := range remaining {
		e1 := node1.Rect.Enlargement(entry.Rect)
		e2 := node2.Rect.Enlargement(entry.Rect)
		if e1 < e2 {
			node1.Entries = append(node1.Entries, entry)
			node1.Rect = node1.Rect.Union(entry.Rect)
		} else {
			node2.Entries = append(node2.Entries, entry)
			node2.Rect = node2.Rect.Union(entry.Rect)
		}
	}

	if n.Parent == nil {
		// New root
		t.Root = &RNode{
			Children: []*RNode{node1, node2},
			Rect:     node1.Rect.Union(node2.Rect),
		}
		node1.Parent = t.Root
		node2.Parent = t.Root
	} else {
		// Update parent
		n.Parent.Entries = append(n.Parent.Entries, &Entry{Child: node1, Rect: node1.Rect})
		n.Parent.Entries = append(n.Parent.Entries, &Entry{Child: node2, Rect: node2.Rect})
		// remove old entry
		for i, e := range n.Parent.Entries {
			if e.Child == n {
				n.Parent.Entries = append(n.Parent.Entries[:i], n.Parent.Entries[i+1:]...)
				break
			}
		}
		if len(n.Parent.Entries) > 4 {
			t.splitNode(n.Parent)
		} else {
			t.adjustTree(n.Parent)
		}
	}
}

func (t *RTree) adjustTree(n *RNode) {
	if n == t.Root {
		return
	}

	parent := n.Parent
	parent.Rect = parent.Children[0].Rect
	for i := 1; i < len(parent.Children); i++ {
		parent.Rect = parent.Rect.Union(parent.Children[i].Rect)
	}

	t.adjustTree(parent)
}

func (t *RTree) Search(rect *city.Rect) []Spatial {
	return t.search(t.Root, rect)
}

func (t *RTree) search(n *RNode, rect *city.Rect) []Spatial {
	var results []Spatial
	if n.Leaf {
		for _, entry := range n.Entries {
			if entry.Rect.Intersects(rect) {
				results = append(results, entry.Obj)
			}
		}
	} else {
		for _, child := range n.Children {
			if child.Rect.Intersects(rect) {
				results = append(results, t.search(child, rect)...)
			}
		}
	}
	return results
}
