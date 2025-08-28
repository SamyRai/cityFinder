package city

import (
	"github.com/dhconnelly/rtreego"
	"math"
)

type City struct {
	Latitude  float64
	Longitude float64
	Name      string
	Country   string
}

type SpatialCity struct {
	City
	Rect rtreego.Rect
}

func (sc *SpatialCity) Bounds() rtreego.Rect {
	return sc.Rect
}

const earthRadiusKm = 6371

// HaversineDistance calculates the distance between two points on Earth.
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := toRadians(lat2 - lat1)
	dLon := toRadians(lon2 - lon1)
	lat1Rad := toRadians(lat1)
	lat2Rad := toRadians(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

func toRadians(deg float64) float64 {
	return deg * (math.Pi / 180)
}
