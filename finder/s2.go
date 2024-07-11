package finder

import (
	"cityFinder/city"
	"cityFinder/dataloader"
	"github.com/golang/geo/s2"
	"math"
)

type S2Finder struct {
	index      *s2.RegionCoverer
	data       map[s2.CellID]city.SpatialCity
	postalCode map[string]dataloader.PostalCodeEntry
}

func BuildS2Index(cities []city.SpatialCity, postalCodes map[string]dataloader.PostalCodeEntry) *S2Finder {
	index := &s2.RegionCoverer{
		MinLevel: 15,
		MaxLevel: 15,
		MaxCells: 8,
	}

	data := make(map[s2.CellID]city.SpatialCity)
	for _, city := range cities {
		cellID := s2.CellIDFromLatLng(s2.LatLngFromDegrees(city.Latitude, city.Longitude)).Parent(15)
		data[cellID] = city
	}

	return &S2Finder{index: index, data: data, postalCode: postalCodes}
}

func (f *S2Finder) FindNearestCity(lat, lon float64) *city.City {
	point := s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lon))
	cap := s2.CapFromPoint(point)
	covering := f.index.Covering(cap)

	minDist := float64(math.MaxFloat64)
	var nearestCity *city.City
	for _, cellID := range covering {
		if city, ok := f.data[cellID]; ok {
			dist := euclideanDistance(lat, lon, city.Latitude, city.Longitude)
			if dist < minDist {
				minDist = dist
				nearestCity = &city.City
			}
		}
	}
	return nearestCity
}

func euclideanDistance(lat1, lon1, lat2, lon2 float64) float64 {
	return math.Sqrt(math.Pow(lat1-lat2, 2) + math.Pow(lon1-lon2, 2))
}

func (f *S2Finder) FindCoordinatesByName(name string) *city.City {
	// Implement this if needed for completeness
	return nil
}

func (f *S2Finder) FindCityByPostalCode(postalCode string) *city.City {
	entry, exists := f.postalCode[postalCode]
	if !exists {
		return nil
	}
	return f.FindNearestCity(entry.Latitude, entry.Longitude)
}
