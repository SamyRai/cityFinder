package coordinates

import (
	"github.com/SamyRai/cityFinder/lib/city"
	"github.com/SamyRai/cityFinder/lib/dataLoader"
	"github.com/cheggaaa/pb/v3"
	"github.com/dhconnelly/rtreego"
)

type RTreeFinder struct {
	tree       *rtreego.Rtree
	nameIndex  map[string]*city.City
	postalCode map[string]dataLoader.PostalCodeEntry
}

func BuildRTree(cities []city.SpatialCity, postalCodes map[string]dataLoader.PostalCodeEntry) *RTreeFinder {
	rtree := rtreego.NewTree(2, 25, 50)
	bar := pb.Full.Start(len(cities))
	defer bar.Finish()
	nameIndex := make(map[string]*city.City)

	for _, city := range cities {
		rtree.Insert(&city)
		nameIndex[city.Name] = &city.City
		bar.Increment()
	}
	return &RTreeFinder{tree: rtree, nameIndex: nameIndex, postalCode: postalCodes}
}

func (f *RTreeFinder) NearestPlace(lat, lon float64) *city.City {
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
		spatialCityAsPoint := rtreego.Point{spatialCity.Longitude, spatialCity.Latitude}
		distance := city.EuclideanDistance(point, spatialCityAsPoint)
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
