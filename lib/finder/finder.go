package finder

import (
	"github.com/SamyRai/cityFinder/lib/city"
	"github.com/SamyRai/cityFinder/lib/config"
	"github.com/SamyRai/cityFinder/lib/dataLoader"
	"github.com/SamyRai/cityFinder/lib/finder/coordinates"
	"github.com/SamyRai/cityFinder/lib/finder/name"
	"github.com/SamyRai/cityFinder/lib/finder/postalCode"
)

// Finder struct embeds all individual finders
type Finder struct {
	S2Finder         *coordinates.S2Finder
	NameFinder       *name.Finder
	PostalCodeFinder *postalCode.Finder
}

// NewFinder creates a new Finder instance
func NewFinder(cities []city.SpatialCity, s2Config *config.S2, postalCodes map[string]map[string]dataLoader.PostalCodeEntry) *Finder {

	s2Finder, err := coordinates.NewS2Finder(s2Config)
	if err != nil {
		panic(err)
	}

	nameFinder := name.NewNameFinder()
	postalCodeFinder := postalCode.NewPostalCodeFinder()

	for _, spatialCity := range cities {
		nameFinder.AddCity(spatialCity)
	}

	for _, postalCodeEntries := range postalCodes {
		for _, entry := range postalCodeEntries {
			postalCodeFinder.AddPostalCode(entry)
		}
	}

	return &Finder{
		S2Finder:         s2Finder,
		NameFinder:       nameFinder,
		PostalCodeFinder: postalCodeFinder,
	}
}

// FindCityByPostalCode wraps the PostalCodeFinder method
func (f *Finder) FindCityByPostalCode(postalCode, countryCode string) *city.City {
	return f.PostalCodeFinder.CityByPostalCode(postalCode, countryCode)
}

// FindCityByName wraps the NameFinder method
func (f *Finder) FindCityByName(name, countryCode string) *city.City {
	return f.NameFinder.CityByName(name, countryCode)
}

// FindNearestCity wraps the S2Finder method
func (f *Finder) FindNearestCity(lat, lon float64) (*city.City, float64, error) {
	c := f.S2Finder.NearestPlace(lat, lon)
	if c == nil {
		return nil, 0, nil
	}
	return c, 0, nil
}
