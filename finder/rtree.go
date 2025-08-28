package finder

import (
	"github.com/SamyRai/cityFinder/city"
	"github.com/SamyRai/cityFinder/dataloader"
	"github.com/cheggaaa/pb/v3"
	"github.com/dhconnelly/rtreego"
)

type RTreeFinder struct {
	tree       *rtreego.Rtree
	nameIndex  map[string][]*city.City
	postalCode map[string]dataloader.PostalCodeEntry
}

func BuildRTree(cities []city.SpatialCity, postalCodes map[string]dataloader.PostalCodeEntry) *RTreeFinder {
	rtree := rtreego.NewTree(2, 25, 50)
	bar := pb.Full.Start(len(cities))
	defer bar.Finish()
	nameIndex := make(map[string][]*city.City)

	for _, c := range cities {
		cityCopy := c // Create a new variable to avoid capturing the loop variable in a closure.
		rtree.Insert(&cityCopy)
		nameIndex[cityCopy.Name] = append(nameIndex[cityCopy.Name], &cityCopy.City)
		bar.Increment()
	}
	return &RTreeFinder{tree: rtree, nameIndex: nameIndex, postalCode: postalCodes}
}

func (f *RTreeFinder) FindNearestCity(lat, lon float64) *city.City {
	point := rtreego.Point{lon, lat}
	rect, _ := rtreego.NewRect(point, []float64{0.1, 0.1}) // Start with a larger search area
	results := f.tree.SearchIntersect(rect)

	minDistance := float64(1<<63 - 1)
	var nearestCity *city.City
	rectSize := 0.1
	for len(results) == 0 { // If no results, expand the search area
		rect, _ = rtreego.NewRect(point, []float64{rectSize, rectSize})
		results = f.tree.SearchIntersect(rect)
		rectSize += 0.1
		if rectSize > 1 {
			break
		}
	}

	for _, item := range results {
		spatialCity := item.(*city.SpatialCity)
		distance := city.HaversineDistance(lat, lon, spatialCity.Latitude, spatialCity.Longitude)
		if distance < minDistance {
			minDistance = distance
			nearestCity = &spatialCity.City
		}
	}
	return nearestCity
}

func (f *RTreeFinder) FindCoordinatesByName(name string) []*city.City {
	if cities, exists := f.nameIndex[name]; exists {
		return cities
	}
	return nil
}

func (f *RTreeFinder) FindCityByPostalCode(postalCode string) *city.City {
	entry, exists := f.postalCode[postalCode]
	if !exists {
		return nil
	}
	return f.FindNearestCity(entry.Latitude, entry.Longitude)
}
