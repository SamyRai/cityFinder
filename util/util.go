package util

import (
	"os"
	"path/filepath"
)

// FindProjectRoot finds the project root directory by looking for a specific file or directory that should be at the root.
func FindProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			break
		}
		dir = parentDir
	}

	return "", os.ErrNotExist
}

// LevenshteinDistance calculates the Levenshtein distance between two strings
func LevenshteinDistance(a, b string) int {
	al := len(a)
	bl := len(b)
	if al == 0 {
		return bl
	}
	if bl == 0 {
		return al
	}

	d := make([][]int, al+1)
	for i := range d {
		d[i] = make([]int, bl+1)
	}

	for i := 0; i <= al; i++ {
		d[i][0] = i
	}
	for j := 0; j <= bl; j++ {
		d[0][j] = j
	}

	for i := 1; i <= al; i++ {
		for j := 1; j <= bl; j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			d[i][j] = min(
				d[i-1][j]+1,      // deletion
				d[i][j-1]+1,      // insertion
				d[i-1][j-1]+cost, // substitution
			)
		}
	}

	return d[al][bl]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// BKTree is a data structure for fast fuzzy string matching
type BKTree struct {
	Root *bkNode
}

type bkNode struct {
	Term     string
	Children map[int]*bkNode
}

// NewBKTree creates a new BK-tree
func NewBKTree() *BKTree {
	return &BKTree{}
}

// Add inserts a term into the BK-tree
func (tree *BKTree) Add(term string) {
	if tree.Root == nil {
		tree.Root = &bkNode{Term: term, Children: make(map[int]*bkNode)}
		return
	}
	current := tree.Root
	for {
		distance := LevenshteinDistance(term, current.Term)
		child, exists := current.Children[distance]
		if !exists {
			current.Children[distance] = &bkNode{Term: term, Children: make(map[int]*bkNode)}
			return
		}
		current = child
	}
}

// Search returns terms in the BK-tree within the given distance of the query term
func (tree *BKTree) Search(query string, maxDistance int) []string {
	if tree.Root == nil {
		return nil
	}
	var results []string
	var search func(*bkNode)
	search = func(node *bkNode) {
		distance := LevenshteinDistance(query, node.Term)
		if distance <= maxDistance {
			results = append(results, node.Term)
		}
		for i := max(1, distance-maxDistance); i <= distance+maxDistance; i++ {
			child, exists := node.Children[i]
			if exists {
				search(child)
			}
		}
	}
	search(tree.Root)
	return results
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
