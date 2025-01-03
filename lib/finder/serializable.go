package finder

import (
	"github.com/SamyRai/cityFinder/lib/city"
)

// FromSpatialCity converts city.SpatialCity to SerializableSpatialCity
func FromSpatialCity(sc city.SpatialCity) SerializableSpatialCity {
	return SerializableSpatialCity{
		Latitude:  sc.Latitude,
		Longitude: sc.Longitude,
		Name:      sc.Name,
		Country:   sc.Country,
		Rect:      FromRTreeRect(sc.Rect),
	}
}

// ToSpatialCity converts SerializableSpatialCity to city.SpatialCity
func ToSpatialCity(ssc SerializableSpatialCity) (city.SpatialCity, error) {
	rect, err := ssc.Rect.ToRTreeRect()
	if err != nil {
		return city.SpatialCity{}, err
	}
	return city.SpatialCity{
		City: city.City{
			Latitude:  ssc.Latitude,
			Longitude: ssc.Longitude,
			Name:      ssc.Name,
			Country:   ssc.Country,
		},
		Rect: rect,
	}, nil
}
