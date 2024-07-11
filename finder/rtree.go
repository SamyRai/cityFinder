package finder

import (
	"cityFinder/city"
	"cityFinder/dataloader"
	"github.com/cheggaaa/pb/v3"
	"github.com/dhconnelly/rtreego"
)

type RTreeFinder struct {
	tree       *rtreego.Rtree
	postalCode map[string]dataloader.PostalCodeEntry
}

func BuildRTree(cities []city.SpatialCity, postalCodes map[string]dataloader.PostalCodeEntry) *RTreeFinder {
	rtree := rtreego.NewTree(2, 25, 50)
	bar := pb.Full.Start(len(cities))
	defer bar.Finish()

	for _, city := range cities {
		rtree.Insert(&city)
		bar.Increment()
	}
	return &RTreeFinder{tree: rtree, postalCode: postalCodes}
}

func (f *RTreeFinder) FindNearestCity(lat, lon float64) *city.City {
	point := rtreego.Point{lon, lat}
	rect, _ := rtreego.NewRect(point, []float64{0.00001, 0.00001})
	results := f.tree.SearchIntersect(rect)

	bar := pb.Full.Start(len(results))
	defer bar.Finish()

	minDistance := float64(1<<63 - 1)
	var nearestCity *city.City

	for _, item := range results {
		spatialCity := item.(*city.SpatialCity)
		spatialCityAsPoint := rtreego.Point{spatialCity.Longitude, spatialCity.Latitude}
		distance := city.EuclideanDistance(point, spatialCityAsPoint)
		if distance < minDistance {
			minDistance = distance
			nearestCity = &spatialCity.City
		}
		bar.Increment()
	}
	return nearestCity
}

func (f *RTreeFinder) FindCoordinatesByName(name string) *city.City {
	// Implement this if needed for completeness
	return nil
}

func (f *RTreeFinder) FindCityByPostalCode(postalCode string) *city.City {
	entry, exists := f.postalCode[postalCode]
	if !exists {
		return nil
	}
	return f.FindNearestCity(entry.Latitude, entry.Longitude)
}
