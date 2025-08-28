package finder

import (
	"errors"
	"github.com/SamyRai/cityFinder/city"
)

// Finder interface that each finder should implement
type Finder interface {
	FindNearestCity(lat, lon float64) (*city.City, float64, error)
	FindCoordinatesByName(name string) []*city.City
	FindCityByPostalCode(postalCode string) []*city.City
	Close() error
}

// Indexer interface that each finder's indexer should implement
type Indexer interface {
	Build(cities <-chan city.SpatialCity) error
	Serialize(path string) error
	Finder
}

var (
	ErrOutOfRange      = errors.New("lat/lon out of range")
	ErrNoResults       = errors.New("no nearby city found")
	ErrCorruptMeta     = errors.New("corrupt or mismatched meta vs data")
	ErrIndexOutOfRange = errors.New("shape edge id out of range")
)

// Meta holds the metadata for any finder index.
type Meta struct {
	FinderType string
	FinderMeta interface{}
}
