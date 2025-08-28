package city

import (
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
