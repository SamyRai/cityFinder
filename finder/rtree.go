package finder

import (
	"encoding/gob"
	"math"
	"os"

	"github.com/SamyRai/cityFinder/city"
	"github.com/dhconnelly/rtreego"
)

type RTreeFinder struct {
	tree *rtreego.Rtree
	*CityReader
}

type RTreeSpatial struct {
	Rect   SerializableRect
	CityID int
}

func (s *RTreeSpatial) Bounds() rtreego.Rect {
	r, _ := s.Rect.ToRTreeRect()
	return r
}

type RTreeMeta struct {
	Spatials    []RTreeSpatial
	CityOffsets []int64
	CityLengths []int64
}

func DeserializeRTree(metaPath, citiesPath string) (*RTreeFinder, error) {
	metaFile, err := os.Open(metaPath)
	if err != nil {
		return nil, err
	}
	defer metaFile.Close()

	var meta RTreeMeta
	if err := gob.NewDecoder(metaFile).Decode(&meta); err != nil {
		return nil, err
	}

	tree := rtreego.NewTree(2, 25, 50)
	for _, s := range meta.Spatials {
		tree.Insert(&s)
	}

	cityReader, err := NewCityReader(citiesPath, meta.CityOffsets, meta.CityLengths)
	if err != nil {
		return nil, err
	}

	return &RTreeFinder{
		tree:       tree,
		CityReader: cityReader,
	}, nil
}

func (f *RTreeFinder) FindNearestCity(lat, lon float64) (*city.City, float64, error) {
	p := rtreego.Point{lon, lat}
	searchRect, _ := rtreego.NewRect(p, []float64{1.0, 1.0})
	results := f.tree.SearchIntersect(searchRect)

	if len(results) == 0 {
		return nil, 0, ErrNoResults
	}

	var nearestCity *city.City
	minDist := math.MaxFloat64

	for _, item := range results {
		s := item.(*RTreeSpatial)
		c, err := f.ReadCityAt(s.CityID)
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

func (f *RTreeFinder) FindCoordinatesByName(name string) []*city.City {
	return nil
}

func (f *RTreeFinder) FindCityByPostalCode(postalCode string) []*city.City {
	return nil
}
