package coordinates

import (
	"github.com/SamyRai/cityFinder/lib/city"
	"github.com/SamyRai/cityFinder/lib/dataLoader"
	"github.com/dhconnelly/rtreego"
	"github.com/mmcloughlin/geohash"
)

type GeoHashFinder struct {
	data       map[string][]city.SpatialCity
	postalCode map[string]dataLoader.PostalCodeEntry
}

func BuildGeoHashIndex(cities []city.SpatialCity, precision uint, postalCodes map[string]dataLoader.PostalCodeEntry) *GeoHashFinder {
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

func (f *GeoHashFinder) NearestPlace(lat, lon float64) *city.City {
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

func (f *GeoHashFinder) FindCityByPostalCode(postalCode string) *city.City {
	entry, exists := f.postalCode[postalCode]
	if !exists {
		return nil
	}
	return f.NearestPlace(entry.Latitude, entry.Longitude)
}
