package finder

import (
	"encoding/gob"
	"fmt"
	"math"
	"os"
	"sync"

	"github.com/SamyRai/cityFinder/city"
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

// S2Meta holds the metadata for the S2 index.
type S2Meta struct {
	CityOffsets  []int64
	CityLengths  []int64
	PointOffsets []int64
}

// S2Finder finds the nearest city using the S2 geometry library.
type S2Finder struct {
	index          *s2.ShapeIndex
	streamingShape *StreamingPointVector
	*CityReader
	query *s2.EdgeQuery
	mu    sync.Mutex
}

// DeserializeS2 loads an S2Finder from the serialized data files.
func DeserializeS2(metaPath, citiesPath, pointsPath string) (*S2Finder, error) {
	metaFile, err := os.Open(metaPath)
	if err != nil {
		return nil, err
	}
	defer metaFile.Close()

	var meta S2Meta
	if err := gob.NewDecoder(metaFile).Decode(&meta); err != nil {
		return nil, err
	}

	streamingShape, err := NewStreamingPointVector(pointsPath, meta.PointOffsets)
	if err != nil {
		return nil, err
	}

	index := s2.NewShapeIndex()
	index.Add(streamingShape)

	cityReader, err := NewCityReader(citiesPath, meta.CityOffsets, meta.CityLengths)
	if err != nil {
		streamingShape.Close()
		return nil, err
	}

	opts := s2.NewClosestEdgeQueryOptions().MaxResults(1).IncludeInteriors(false)
	query := s2.NewClosestEdgeQuery(index, opts)

	return &S2Finder{
		index:          index,
		streamingShape: streamingShape,
		CityReader:     cityReader,
		query:          query,
	}, nil
}

// Close closes the file handles held by the S2Finder.
func (f *S2Finder) Close() error {
	f.streamingShape.Close()
	return f.CityReader.Close()
}

// FindNearestCity returns the closest city to (lat, lon), along with the
// great-circle distance (in kilometers). Returns error if none found.
func (f *S2Finder) FindNearestCity(lat, lon float64) (*city.City, float64, error) {
	if math.Abs(lat) > 90 || math.Abs(lon) > 180 || math.IsNaN(lat) || math.IsNaN(lon) {
		return nil, 0, ErrOutOfRange
	}

	targetPoint := s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lon))

	f.mu.Lock()
	defer f.mu.Unlock()

	f.query.Reset()
	target := s2.NewMinDistanceToPointTarget(targetPoint)
	results := f.query.FindEdges(target)
	if len(results) == 0 {
		return nil, 0, ErrNoResults
	}

	r := results[0]
	edgeID := r.EdgeID()
	if edgeID < 0 || int(edgeID) >= len(f.CityReader.offsets) {
		return nil, 0, ErrIndexOutOfRange
	}

	angle := r.Distance().Angle()
	km := angle.Radians() * 6371.0088

	c, err := f.ReadCityAt(int(edgeID))
	if err != nil {
		return nil, 0, fmt.Errorf("read city: %w", err)
	}
	return c, km, nil
}

// FindCoordinatesByName is not implemented for this finder.
func (f *S2Finder) FindCoordinatesByName(name string) []*city.City {
	return nil
}

// FindCityByPostalCode is not implemented for this finder.
func (f *S2Finder) FindCityByPostalCode(postalCode string) []*city.City {
	return nil
}

// HaversineDistance computes the distance between two points on the Earth's surface.
func HaversineDistance(p1, p2 s2.Point) s1.Angle {
	return p1.Distance(p2)
}
