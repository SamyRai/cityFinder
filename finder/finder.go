package finder

import "github.com/SamyRai/cityFinder/city"

// Finder interface that each finder should implement
type Finder interface {
	FindNearestCity(lat, lon float64) *city.City
	FindCoordinatesByName(name string) *city.City
	FindCityByPostalCode(postalCode string) *city.City
}
