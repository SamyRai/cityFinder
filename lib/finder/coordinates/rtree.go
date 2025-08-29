package coordinates

import (
	"github.com/SamyRai/cityFinder/lib/city"
	"github.com/SamyRai/cityFinder/lib/dataLoader"
	"github.com/cheggaaa/pb/v3"
)

type RTreeFinder struct {
	tree       *RTree
	nameIndex  map[string]*city.City
	postalCode map[string]dataLoader.PostalCodeEntry
}

func BuildRTree(cities []city.SpatialCity, postalCodes map[string]dataLoader.PostalCodeEntry) *RTreeFinder {
	rtree := NewRTree(25)
	bar := pb.Full.Start(len(cities))
	defer bar.Finish()
	nameIndex := make(map[string]*city.City)

	for i := range cities {
		rtree.Insert(&cities[i])
		nameIndex[cities[i].Name] = &cities[i].City
		bar.Increment()
	}
	return &RTreeFinder{tree: rtree, nameIndex: nameIndex, postalCode: postalCodes}
}

func (f *RTreeFinder) NearestPlace(lat, lon float64) *city.City {
	rect := &city.Rect{
		Min: []float64{lon - 0.1, lat - 0.1},
		Max: []float64{lon + 0.1, lat + 0.1},
	}
	results := f.tree.Search(rect)

	minDistance := float64(1<<63 - 1)
	var nearestCity *city.City

	for _, item := range results {
		spatialCity := item.(*city.SpatialCity)
		distance := city.EuclideanDistance([]float64{lon, lat}, []float64{spatialCity.Longitude, spatialCity.Latitude})
		if distance < minDistance {
			minDistance = distance
			nearestCity = &spatialCity.City
		}
	}
	return nearestCity
}

func (f *RTreeFinder) FindCityByPostalCode(postalCode string) *city.City {
	entry, exists := f.postalCode[postalCode]
	if !exists {
		return nil
	}
	return f.NearestPlace(entry.Latitude, entry.Longitude)
}
