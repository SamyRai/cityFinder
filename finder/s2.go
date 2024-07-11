package finder

import (
	"encoding/gob"
	"github.com/SamyRai/cityFinder/city"
	"github.com/SamyRai/cityFinder/dataloader"
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	"math"
	"os"
)

type S2Finder struct {
	Index      *s2.RegionCoverer
	Data       map[s2.CellID]SerializableSpatialCity
	PostalCode map[string]dataloader.PostalCodeEntry
	NameIndex  map[string]*city.City
}

func BuildS2Index(cities []city.SpatialCity, postalCodes map[string]dataloader.PostalCodeEntry) *S2Finder {
	index := &s2.RegionCoverer{
		MinLevel: 15,
		MaxLevel: 15,
		MaxCells: 8,
	}

	data := make(map[s2.CellID]SerializableSpatialCity)
	for _, city := range cities {
		cellID := s2.CellIDFromLatLng(s2.LatLngFromDegrees(city.Latitude, city.Longitude)).Parent(15)
		data[cellID] = FromSpatialCity(city)
	}

	return &S2Finder{Index: index, Data: data, PostalCode: postalCodes}
}

func (f *S2Finder) FindNearestCity(lat, lon float64) *city.City {
	point := s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lon))
	cap := s2.CapFromPoint(point).Expanded(s1.Angle(1e-5))
	covering := f.Index.Covering(cap)

	minDist := float64(math.MaxFloat64)
	var nearestCity *city.City
	for _, cellID := range covering {
		if ssc, ok := f.Data[cellID]; ok {
			sc, err := ToSpatialCity(ssc)
			if err != nil {
				continue
			}
			dist := euclideanDistance(lat, lon, sc.Latitude, sc.Longitude)
			if dist < minDist {
				minDist = dist
				nearestCity = &sc.City
			}
		}
	}
	return nearestCity
}

func euclideanDistance(lat1, lon1, lat2, lon2 float64) float64 {
	return math.Sqrt(math.Pow(lat1-lat2, 2) + math.Pow(lon1-lon2, 2))
}

func (f *S2Finder) FindCoordinatesByName(name string) *city.City {
	if city, exists := f.NameIndex[name]; exists {
		return city
	}
	return nil
}

func (f *S2Finder) FindCityByPostalCode(postalCode string) *city.City {
	entry, exists := f.PostalCode[postalCode]
	if !exists {
		return nil
	}
	return f.FindNearestCity(entry.Latitude, entry.Longitude)
}

// SerializeIndex saves the index to a file
func (f *S2Finder) SerializeIndex(filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(f)
	return err
}

// DeserializeIndex loads the index from a file
func DeserializeIndex(filepath string) (*S2Finder, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	var finder S2Finder
	err = decoder.Decode(&finder)
	if err != nil {
		return nil, err
	}
	return &finder, nil
}

func init() {
	gob.Register(&s2.RegionCoverer{})
	gob.Register(s2.CellID(0))
	gob.Register(SerializableSpatialCity{})
	gob.Register(SerializableRect{})
}
