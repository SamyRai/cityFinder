package coordinates

import (
	"github.com/SamyRai/cityFinder/lib/city"
	"github.com/SamyRai/cityFinder/lib/dataLoader"
	"math"
)

type KDTreeFinder struct {
	tree       *KDTree
	postalCode map[string]dataLoader.PostalCodeEntry
	index      map[string]*city.City
}

type kdPoint struct {
	Coordinates []float64
	City        *city.City
}

func (p kdPoint) Dimensions() int {
	return len(p.Coordinates)
}

func (p kdPoint) Dimension(i int) float64 {
	return p.Coordinates[i]
}

func (p kdPoint) Distance(q Point) float64 {
	other := q.(kdPoint)
	dist := 0.0
	for i := 0; i < len(p.Coordinates); i++ {
		diff := p.Coordinates[i] - other.Coordinates[i]
		dist += diff * diff
	}
	return math.Sqrt(dist)
}

func BuildKDTree(cities []city.SpatialCity, postalCodes map[string]dataLoader.PostalCodeEntry) *KDTreeFinder {
	points := make([]Point, len(cities))
	nameIndex := make(map[string]*city.City)
	for i, c := range cities {
		points[i] = kdPoint{
			Coordinates: []float64{c.Longitude, c.Latitude},
			City:        &c.City,
		}
		nameIndex[c.Name] = &c.City
	}
	tree := NewKDTree(points)
	return &KDTreeFinder{tree: tree, postalCode: postalCodes, index: nameIndex}
}

func (f *KDTreeFinder) NearestPlace(lat, lon float64) *city.City {
	target := kdPoint{
		Coordinates: []float64{lon, lat},
	}
	nearest := f.tree.Nearest(target)
	if nearest != nil {
		return nearest.(kdPoint).City
	}
	return nil
}

func (f *KDTreeFinder) FindCityByPostalCode(postalCode string) *city.City {
	entry, exists := f.postalCode[postalCode]
	if !exists {
		return nil
	}
	return f.NearestPlace(entry.Latitude, entry.Longitude)
}
