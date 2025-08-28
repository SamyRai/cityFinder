package finder

import (
	"encoding/gob"
	"math"
	"os"

	"github.com/SamyRai/cityFinder/city"
	"github.com/kyroy/kdtree"
)

type KDTreeFinder struct {
	tree *kdtree.KDTree
	*CityReader
}

type KDTreePoint struct {
	Coordinates []float64
	CityID      int
}

func (p KDTreePoint) Dimensions() int {
	return len(p.Coordinates)
}

func (p KDTreePoint) Dimension(i int) float64 {
	return p.Coordinates[i]
}

func (p KDTreePoint) Distance(other kdtree.Point) float64 {
	o := other.(KDTreePoint)
	dx := p.Coordinates[0] - o.Coordinates[0]
	dy := p.Coordinates[1] - o.Coordinates[1]
	return dx*dx + dy*dy
}

type KDTreeMeta struct {
	Points      []KDTreePoint
	CityOffsets []int64
	CityLengths []int64
}

func DeserializeKDTree(metaPath, citiesPath string) (*KDTreeFinder, error) {
	metaFile, err := os.Open(metaPath)
	if err != nil {
		return nil, err
	}
	defer metaFile.Close()

	var meta KDTreeMeta
	if err := gob.NewDecoder(metaFile).Decode(&meta); err != nil {
		return nil, err
	}

	points := make([]kdtree.Point, len(meta.Points))
	for i, p := range meta.Points {
		points[i] = p
	}
	tree := kdtree.New(points)

	cityReader, err := NewCityReader(citiesPath, meta.CityOffsets, meta.CityLengths)
	if err != nil {
		return nil, err
	}

	return &KDTreeFinder{
		tree:       tree,
		CityReader: cityReader,
	}, nil
}

func (f *KDTreeFinder) FindNearestCity(lat, lon float64) (*city.City, float64, error) {
	target := KDTreePoint{Coordinates: []float64{lon, lat}}
	candidates := f.tree.KNN(target, 100)

	if len(candidates) == 0 {
		return nil, 0, ErrNoResults
	}

	var nearestCity *city.City
	minDist := math.MaxFloat64

	for _, candidate := range candidates {
		p := candidate.(KDTreePoint)
		c, err := f.ReadCityAt(p.CityID)
		if err != nil {
			continue
		}
		dist := city.HaversineDistance(lat, lon, c.Latitude, c.Longitude)
		if dist < minDist {
			minDist = dist
			nearestCity = c
		}
	}

	if nearestCity == nil {
		return nil, 0, ErrNoResults
	}

	return nearestCity, minDist, nil
}

func (f *KDTreeFinder) FindCoordinatesByName(name string) []*city.City {
	return nil
}

func (f *KDTreeFinder) FindCityByPostalCode(postalCode string) []*city.City {
	return nil
}
