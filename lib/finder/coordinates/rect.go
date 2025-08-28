package coordinates

import (
	"github.com/dhconnelly/rtreego"
)

// SerializableSpatialCity is a custom type for serializing city.SpatialCity
type SerializableSpatialCity struct {
	Latitude  float64
	Longitude float64
	Name      string
	Country   string
	Rect      SerializableRect
}

// SerializableRect is a custom type for serializing rtreego.Rect
type SerializableRect struct {
	Point []float64
	Sizes []float64
}

// ToRTreeRect converts a SerializableRect to rtreego.Rect
func (sr *SerializableRect) ToRTreeRect() (rtreego.Rect, error) {
	return rtreego.NewRect(sr.Point, sr.Sizes)
}

// FromRTreeRect converts a rtreego.Rect to SerializableRect
func FromRTreeRect(rect rtreego.Rect) SerializableRect {
	point := make([]float64, int(rect.Size()))
	sizes := make([]float64, int(rect.Size()))

	for i := 0; i < len(point); i++ {
		point[i] = rect.PointCoord(i)
		sizes[i] = rect.LengthsCoord(i)
	}

	return SerializableRect{
		Point: point,
		Sizes: sizes,
	}
}
