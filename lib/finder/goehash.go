package finder

import (
	"github.com/SamyRai/cityFinder/lib/city"
	"github.com/SamyRai/cityFinder/lib/dataloader"
	"github.com/dhconnelly/rtreego"
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
	closest := city.SpatialCity{}
	minDist := float64(1<<63 - 1)
	for key, cities := range f.data {
		if key[:5] == hash[:5] {
			for _, c := range cities {
				dist := city.EuclideanDistance(rtreego.Point{lon, lat}, rtreego.Point{c.Longitude, c.Latitude})
				if dist < minDist {
					minDist = dist
					closest = c
				}
			}
		}
	}
	if closest.City.Name == "" {
		return nil
	}
	return &closest.City
}

func (f *GeoHashFinder) FindCoordinatesByName(name string) *city.City {
	// Implement this if needed for completeness
	return nil
}

func (f *GeoHashFinder) FindCityByPostalCode(postalCode string) *city.City {
	entry, exists := f.postalCode[postalCode]
	if !exists {
		return nil
	}
	return f.FindNearestCity(entry.Latitude, entry.Longitude)
}
