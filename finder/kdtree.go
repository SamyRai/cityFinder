package finder

import (
	"cityFinder/city"
	"cityFinder/dataloader"
	"github.com/kyroy/kdtree"
	"math"
)

type KDTreeFinder struct {
	tree       *kdtree.KDTree
	postalCode map[string]dataloader.PostalCodeEntry
}

type Point struct {
	Coordinates []float64
	City        *city.City
}

func (p Point) Dimensions() int {
	return len(p.Coordinates)
}

func (p Point) Dimension(i int) float64 {
	return p.Coordinates[i]
}

func (p Point) Distance(q kdtree.Point) float64 {
	other := q.(Point)
	dist := 0.0
	for i := 0; i < len(p.Coordinates); i++ {
		diff := p.Coordinates[i] - other.Coordinates[i]
		dist += diff * diff
	}
	return math.Sqrt(dist)
}

func BuildKDTree(cities []city.SpatialCity, postalCodes map[string]dataloader.PostalCodeEntry) *KDTreeFinder {
	points := make([]kdtree.Point, len(cities))
	for i, city := range cities {
		points[i] = Point{
			Coordinates: []float64{city.Longitude, city.Latitude},
			City:        &city.City,
		}
	}
	tree := kdtree.New(points)
	return &KDTreeFinder{tree: tree, postalCode: postalCodes}
}

func (f *KDTreeFinder) FindNearestCity(lat, lon float64) *city.City {
	target := Point{
		Coordinates: []float64{lon, lat},
	}
	nearest := f.tree.KNN(target, 1)
	if len(nearest) > 0 {
		return nearest[0].(Point).City
	}
	return nil
}

func (f *KDTreeFinder) FindCoordinatesByName(name string) *city.City {
	// Implement this if needed for completeness
	return nil
}

func (f *KDTreeFinder) FindCityByPostalCode(postalCode string) *city.City {
	entry, exists := f.postalCode[postalCode]
	if !exists {
		return nil
	}
	return f.FindNearestCity(entry.Latitude, entry.Longitude)
}
