package coordinates

import (
	"math"
	"sort"
)

type Point interface {
	Dimensions() int
	Dimension(i int) float64
	Distance(p Point) float64
}

type KDTree struct {
	Root *KDNode
}

type KDNode struct {
	Point      Point
	Left, Right *KDNode
}

func NewKDTree(points []Point) *KDTree {
	if len(points) == 0 {
		return &KDTree{}
	}
	return &KDTree{Root: build(points, 0)}
}

func build(points []Point, depth int) *KDNode {
	if len(points) == 0 {
		return nil
	}

	axis := depth % points[0].Dimensions()
	median := len(points) / 2

	sort.Slice(points, func(i, j int) bool {
		return points[i].Dimension(axis) < points[j].Dimension(axis)
	})

	return &KDNode{
		Point: points[median],
		Left:  build(points[:median], depth+1),
		Right: build(points[median+1:], depth+1),
	}
}

func (t *KDTree) Nearest(p Point) Point {
	if t.Root == nil {
		return nil
	}
	best, _ := t.nearest(t.Root, p, 0)
	return best
}

func (t *KDTree) nearest(node *KDNode, p Point, depth int) (Point, float64) {
	if node == nil {
		return nil, math.Inf(1)
	}

	axis := depth % p.Dimensions()
	var nextNode, otherNode *KDNode
	if p.Dimension(axis) < node.Point.Dimension(axis) {
		nextNode = node.Left
		otherNode = node.Right
	} else {
		nextNode = node.Right
		otherNode = node.Left
	}

	best, dist := t.nearest(nextNode, p, depth+1)

	if node.Point.Distance(p) < dist {
		best = node.Point
		dist = best.Distance(p)
	}

	if math.Abs(p.Dimension(axis)-node.Point.Dimension(axis)) < dist {
		otherBest, otherDist := t.nearest(otherNode, p, depth+1)
		if otherDist < dist {
			best = otherBest
			dist = otherDist
		}
	}

	return best, dist
}
