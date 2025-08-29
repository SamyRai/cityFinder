package city

import (
	"math"
	"github.com/dhconnelly/rtreego"
)

type City struct {
	Latitude  float64
	Longitude float64
	Name      string
	Country   string
	AltNames  []string
}

type SpatialCity struct {
	City
	Rect rtreego.Rect
}

func (sc *SpatialCity) Bounds() rtreego.Rect {
	return sc.Rect
}

func EuclideanDistance(p1, p2 rtreego.Point) float64 {
	sum := 0.0
	for i := 0; i < len(p1); i++ {
		diff := p1[i] - p2[i]
		sum += diff * diff
	}
	return sum
}

// HaversineDistance calculates the distance between two geographical points in kilometers
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0 // Earth's radius in kilometers

	dLat := toRadians(lat2 - lat1)
	dLon := toRadians(lon2 - lon1)

	a := sin(dLat/2)*sin(dLat/2) +
		cos(toRadians(lat1))*cos(toRadians(lat2))*
			sin(dLon/2)*sin(dLon/2)
	c := 2 * atan2(sqrt(a), sqrt(1-a))

	return R * c
}

// Helper functions for HaversineDistance
func toRadians(deg float64) float64 {
	return deg * (math.Pi / 180.0)
}

func sin(x float64) float64 { return math.Sin(x) }
func cos(x float64) float64 { return math.Cos(x) }
func sqrt(x float64) float64 { return math.Sqrt(x) }
func atan2(y, x float64) float64 { return math.Atan2(y, x) }
