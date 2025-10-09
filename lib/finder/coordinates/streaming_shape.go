package coordinates

import (
	"encoding/gob"
	"os"

	"github.com/golang/geo/s2"
)

// StreamingPointVector is a custom s2.Shape implementation that streams points from a file.
type StreamingPointVector struct {
	s2.PointVector
	file    *os.File
	offsets []int64
}

// NewStreamingPointVector creates a new StreamingPointVector.
func NewStreamingPointVector(path string, offsets []int64) (*StreamingPointVector, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &StreamingPointVector{
		file:    file,
		offsets: offsets,
	}, nil
}

// NumEdges returns the number of points (edges) in the shape.
func (s *StreamingPointVector) NumEdges() int {
	return len(s.offsets)
}

// Edge returns the i-th edge (a degenerate edge representing a point).
func (s *StreamingPointVector) Edge(i int) s2.Edge {
	if i < 0 || i >= len(s.offsets) {
		return s2.Edge{}
	}
	if _, err := s.file.Seek(s.offsets[i], 0); err != nil {
		return s2.Edge{}
	}
	var p s2.Point
	decoder := gob.NewDecoder(s.file)
	if err := decoder.Decode(&p); err != nil {
		return s2.Edge{}
	}
	return s2.Edge{V0: p, V1: p}
}

// ReferencePoint returns a reference point for the shape.
func (s *StreamingPointVector) ReferencePoint() s2.ReferencePoint {
	return s2.ReferencePoint{Contained: false}
}

// NumChains returns the number of chains in the shape.
func (s *StreamingPointVector) NumChains() int {
	return s.NumEdges()
}

// Chain returns the i-th chain.
func (s *StreamingPointVector) Chain(i int) s2.Chain {
	return s2.Chain{Start: i, Length: 1}
}

// ChainEdge returns the edge at the given offset in the given chain.
func (s *StreamingPointVector) ChainEdge(chainID, offset int) s2.Edge {
	return s.Edge(chainID)
}

// ChainPosition returns the chain and offset for the given edge ID.
func (s *StreamingPointVector) ChainPosition(edgeID int) s2.ChainPosition {
	return s2.ChainPosition{ChainID: edgeID, Offset: 0}
}

// Dimension returns the dimension of the geometry (0 for points).
func (s *StreamingPointVector) Dimension() int {
	return 0
}

// IsEmpty returns true if the shape contains no points.
func (s *StreamingPointVector) IsEmpty() bool {
	return len(s.offsets) == 0
}

// IsFull returns true if the shape contains all points on the sphere.
func (s *StreamingPointVector) IsFull() bool {
	return false
}

// Close closes the underlying file.
func (s *StreamingPointVector) Close() error {
	return s.file.Close()
}
