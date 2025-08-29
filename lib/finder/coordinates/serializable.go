package coordinates

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/SamyRai/cityFinder/lib/city"
)

// ErrNoResults is returned when no results are found
var ErrNoResults = errors.New("no results found")

// SerializableSpatialCity is a custom type for serializing city.SpatialCity
type SerializableSpatialCity struct {
	Latitude  float64
	Longitude float64
	Name      string
	Country   string
	Rect      *city.Rect
}

// CityReader provides random access to a gob-encoded file of cities.
type CityReader struct {
	file    *os.File
	offsets []int64
	lengths []int64
}

// NewCityReader creates a new CityReader.
func NewCityReader(path string, offsets, lengths []int64) (*CityReader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &CityReader{
		file:    file,
		offsets: offsets,
		lengths: lengths,
	}, nil
}

// ReadCityAt reads the city record at the given index.
func (r *CityReader) ReadCityAt(i int) (*city.City, error) {
	if i < 0 || i >= len(r.offsets) {
		return nil, fmt.Errorf("index out of range: %d", i)
	}
	offset := r.offsets[i]
	length := r.lengths[i]
	if length <= 0 {
		return nil, fmt.Errorf("invalid record length at %d", i)
	}
	sr := io.NewSectionReader(r.file, offset, length)

	var sc SerializableSpatialCity
	if err := gob.NewDecoder(sr).Decode(&sc); err != nil {
		return nil, fmt.Errorf("gob decode city: %w", err)
	}
	spatial, err := ToSpatialCity(sc)
	if err != nil {
		return nil, fmt.Errorf("to spatial city: %w", err)
	}
	return &spatial.City, nil
}

// Close closes the underlying file.
func (r *CityReader) Close() error {
	return r.file.Close()
}

// FromSpatialCity converts city.SpatialCity to SerializableSpatialCity
func FromSpatialCity(sc city.SpatialCity) SerializableSpatialCity {
	return SerializableSpatialCity{
		Latitude:  sc.Latitude,
		Longitude: sc.Longitude,
		Name:      sc.Name,
		Country:   sc.Country,
		Rect:      sc.Rect,
	}
}

// ToSpatialCity converts SerializableSpatialCity to city.SpatialCity
func ToSpatialCity(ssc SerializableSpatialCity) (city.SpatialCity, error) {
	return city.SpatialCity{
		City: city.City{
			Latitude:  ssc.Latitude,
			Longitude: ssc.Longitude,
			Name:      ssc.Name,
			Country:   ssc.Country,
		},
		Rect: ssc.Rect,
	}, nil
}
