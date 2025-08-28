package finder

import (
	"github.com/SamyRai/cityFinder/city"
	"github.com/SamyRai/cityFinder/dataloader"
	"github.com/kyroy/kdtree"
)

type KDTreeFinder struct {
	tree       *kdtree.KDTree
	postalCode map[string]dataloader.PostalCodeEntry
	index      map[string][]*city.City
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
	// The kdtree library expects a squared Euclidean distance.
	// We can't provide that, but we can provide the Haversine distance.
	// This will work for finding the nearest neighbor, as the order is preserved.
	// Note that the library will take the square root of this value, so the actual distance returned by the library will be incorrect.
	return city.HaversineDistance(p.Coordinates[1], p.Coordinates[0], other.Coordinates[1], other.Coordinates[0])
}

func BuildKDTree(cities []city.SpatialCity, postalCodes map[string]dataloader.PostalCodeEntry) *KDTreeFinder {
	points := make([]kdtree.Point, len(cities))
	nameIndex := make(map[string][]*city.City)
	for i, c := range cities {
		cityCopy := c // Create a new variable to avoid capturing the loop variable in a closure.
		points[i] = Point{
			Coordinates: []float64{cityCopy.Longitude, cityCopy.Latitude},
			City:        &cityCopy.City,
		}
		nameIndex[cityCopy.Name] = append(nameIndex[cityCopy.Name], &cityCopy.City)
	}
	tree := kdtree.New(points)
	return &KDTreeFinder{tree: tree, postalCode: postalCodes, index: nameIndex}
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

func (f *KDTreeFinder) FindCoordinatesByName(name string) []*city.City {
	if cities, exists := f.index[name]; exists {
		return cities
	}
	return nil
}

func (f *KDTreeFinder) FindCityByPostalCode(postalCode string) *city.City {
	entry, exists := f.postalCode[postalCode]
	if !exists {
		return nil
	}
	return f.FindNearestCity(entry.Latitude, entry.Longitude)
}
