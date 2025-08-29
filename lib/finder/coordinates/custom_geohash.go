package coordinates

import (
	"bytes"
	"strings"
)

const (
	base32 = "0123456789bcdefghjkmnpqrstuvwxyz"
)

func Encode(latitude, longitude float64, precision int) string {
	var geohash bytes.Buffer
	var latRange = []float64{-90.0, 90.0}
	var lonRange = []float64{-180.0, 180.0}
	var isEven = true
	var bit, ch int

	for geohash.Len() < precision {
		if isEven {
			mid := (lonRange[0] + lonRange[1]) / 2
			if longitude > mid {
				ch |= 1 << (4 - uint(bit))
				lonRange[0] = mid
			} else {
				lonRange[1] = mid
			}
		} else {
			mid := (latRange[0] + latRange[1]) / 2
			if latitude > mid {
				ch |= 1 << (4 - uint(bit))
				latRange[0] = mid
			} else {
				latRange[1] = mid
			}
		}

		isEven = !isEven
		bit++

		if bit == 5 {
			geohash.WriteByte(base32[ch])
			bit = 0
			ch = 0
		}
	}

	return geohash.String()
}

func Decode(geohash string) (float64, float64, float64, float64) {
	latRange := []float64{-90.0, 90.0}
	lonRange := []float64{-180.0, 180.0}
	isEven := true

	for _, r := range geohash {
		ch := strings.IndexRune(base32, r)
		for i := 0; i < 5; i++ {
			mask := 1 << (4 - uint(i))
			if isEven {
				mid := (lonRange[0] + lonRange[1]) / 2
				if ch&mask != 0 {
					lonRange[0] = mid
				} else {
					lonRange[1] = mid
				}
			} else {
				mid := (latRange[0] + latRange[1]) / 2
				if ch&mask != 0 {
					latRange[0] = mid
				} else {
					latRange[1] = mid
				}
			}
			isEven = !isEven
		}
	}

	lat := (latRange[0] + latRange[1]) / 2
	lon := (lonRange[0] + lonRange[1]) / 2
	latErr := latRange[1] - lat
	lonErr := lonRange[1] - lon

	return lat, lon, latErr, lonErr
}

func Neighbors(geohash string) []string {
	lat, lon, latErr, lonErr := Decode(geohash)
	precision := len(geohash)

	return []string{
		Encode(lat, lon+2*lonErr, precision),
		Encode(lat+2*latErr, lon+2*lonErr, precision),
		Encode(lat+2*latErr, lon, precision),
		Encode(lat+2*latErr, lon-2*lonErr, precision),
		Encode(lat, lon-2*lonErr, precision),
		Encode(lat-2*latErr, lon-2*lonErr, precision),
		Encode(lat-2*latErr, lon, precision),
		Encode(lat-2*latErr, lon+2*lonErr, precision),
	}
}
