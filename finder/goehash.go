package finder

import (
	"github.com/SamyRai/cityFinder/city"
	"github.com/SamyRai/cityFinder/dataloader"
	"github.com/mmcloughlin/geohash"
)

type GeoHashFinder struct {
	data       map[string][]city.SpatialCity
	postalCode map[string]dataloader.PostalCodeEntry
}

func BuildGeoHashIndex(cities []city.SpatialCity, precision uint, postalCodes map[string]dataloader.PostalCodeEntry) *GeoHashFinder {
	index := &GeoHashFinder{
		data:       make(map[string][]city.SpatialCity),
		postalCode: postalCodes,
	}
	for _, city := range cities {
		hash := geohash.EncodeWithPrecision(city.Latitude, city.Longitude, precision)
		index.data[hash] = append(index.data[hash], city)
	}
	return index
}

func (f *GeoHashFinder) FindNearestCity(lat, lon float64) *city.City {
	hash := geohash.EncodeWithPrecision(lat, lon, 12)
	var closest *city.City
	minDist := float64(1<<63 - 1)

	neighbors := geohash.Neighbors(hash)
	searchHashes := append(neighbors, hash)

	for _, h := range searchHashes {
		if cities, ok := f.data[h]; ok {
			for _, c := range cities {
				dist := city.HaversineDistance(lat, lon, c.Latitude, c.Longitude)
				if dist < minDist {
					minDist = dist
					cityCopy := c.City // Create a copy to avoid pointer issues
					closest = &cityCopy
				}
			}
		}
	}
	return closest
}

func (f *GeoHashFinder) FindCoordinatesByName(name string) []*city.City {
	var foundCities []*city.City
	for _, cities := range f.data {
		for _, c := range cities {
			if c.Name == name {
				cityCopy := c.City
				foundCities = append(foundCities, &cityCopy)
			}
		}
	}
	return foundCities
}

func (f *GeoHashFinder) FindCityByPostalCode(postalCode string) *city.City {
	entry, exists := f.postalCode[postalCode]
	if !exists {
		return nil
	}
	return f.FindNearestCity(entry.Latitude, entry.Longitude)
}
