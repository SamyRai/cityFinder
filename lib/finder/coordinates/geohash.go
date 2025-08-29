package coordinates

import (
	"encoding/gob"
	"math"
	"os"

	"github.com/SamyRai/cityFinder/lib/city"
)

type GeoHashFinder struct {
	data map[string][]int
	*CityReader
}

type GeoHashMeta struct {
	Data        map[string][]int
	CityOffsets []int64
	CityLengths []int64
}

func DeserializeGeoHash(metaPath, citiesPath string) (*GeoHashFinder, error) {
	metaFile, err := os.Open(metaPath)
	if err != nil {
		return nil, err
	}
	defer metaFile.Close()

	var meta GeoHashMeta
	if err := gob.NewDecoder(metaFile).Decode(&meta); err != nil {
		return nil, err
	}

	cityReader, err := NewCityReader(citiesPath, meta.CityOffsets, meta.CityLengths)
	if err != nil {
		return nil, err
	}

	return &GeoHashFinder{
		data:       meta.Data,
		CityReader: cityReader,
	}, nil
}

func (f *GeoHashFinder) FindNearestCity(lat, lon float64) (*city.City, float64, error) {
	precision := 12
	hash := Encode(lat, lon, precision)
	var closest *city.City
	minDist := math.MaxFloat64

	neighbors := Neighbors(hash)
	searchHashes := append(neighbors, hash)

	for _, h := range searchHashes {
		if cityIDs, ok := f.data[h]; ok {
			for _, cityID := range cityIDs {
				c, err := f.ReadCityAt(cityID)
				if err != nil {
					continue
				}
				dist := city.HaversineDistance(lat, lon, c.Latitude, c.Longitude)
				if dist < minDist {
					minDist = dist
					closest = c
				}
			}
		}
	}

	if closest == nil {
		return nil, 0, ErrNoResults
	}

	return closest, minDist, nil
}

func (f *GeoHashFinder) FindCoordinatesByName(name string) []*city.City {
	return nil
}

func (f *GeoHashFinder) FindCityByPostalCode(postalCode string) []*city.City {
	return nil
}
