package coordinates

import (
	"encoding/gob"
	"fmt"
	"github.com/SamyRai/cityFinder/lib/city"
	"github.com/SamyRai/cityFinder/lib/config"
	"github.com/cheggaaa/pb/v3"
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	"math"
	"os"
)

// S2Finder is a struct that contains the spatial index and data for cities
type S2Finder struct {
	Index *s2.RegionCoverer                     // S2 region coverer for spatial indexing
	Data  map[s2.CellID]SerializableSpatialCity // Map of cell ID to spatial city data
}

// NewS2Finder creates a new S2Finder instance
func NewS2Finder(cfgS2 *config.S2) (*S2Finder, error) {
	return DeserializeIndex(cfgS2.IndexFile)
}

// BuildIndex creates an S2 spatial index from city and postal code data
func BuildIndex(cities []city.SpatialCity, config *config.S2) (*S2Finder, error) {
	// Initialize the S2 region coverer
	index := &s2.RegionCoverer{
		MinLevel: config.MinLevel,
		MaxLevel: config.MaxLevel,
		MaxCells: config.MaxCells,
	}

	// Create the data map for storing spatial city data
	data := make(map[s2.CellID]SerializableSpatialCity, len(cities))

	bar := pb.Full.Start(len(cities))
	for _, spatialCity := range cities {
		cellID := s2.CellIDFromLatLng(s2.LatLngFromDegrees(spatialCity.Latitude, spatialCity.Longitude)).Parent(15)
		data[cellID] = FromSpatialCity(spatialCity)
		bar.Increment()
	}
	bar.Finish()

	return &S2Finder{Index: index, Data: data}, nil
}

// FindNearestCity finds the nearest city to the given latitude and longitude
func (f *S2Finder) NearestPlace(lat, lon float64) *city.City {
	point := s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lon))
	initialRadius := 1e-5 // Start with a small radius
	maxRadius := 0.1      // Set a maximum radius to avoid infinite loops

	var nearestCity *city.City
	for radius := initialRadius; radius <= maxRadius; radius *= 2 {
		expanded := s2.CapFromPoint(point).Expanded(s1.Angle(radius))
		covering := f.Index.Covering(expanded)

		minDist := math.MaxFloat64
		for _, cellID := range covering {
			if ssc, ok := f.Data[cellID]; ok {
				sc, err := ToSpatialCity(ssc)
				if err != nil {
					continue
				}
				dist := euclideanDistance(lat, lon, sc.Latitude, sc.Longitude)
				if dist < minDist {
					minDist = dist
					nearestCity = &sc.City
				}
			}
		}

		if nearestCity != nil {
			return nearestCity
		}
	}

	return nil
}

// euclideanDistance calculates the Euclidean distance between two points
func euclideanDistance(lat1, lon1, lat2, lon2 float64) float64 {
	return math.Sqrt(math.Pow(lat1-lat2, 2) + math.Pow(lon1-lon2, 2))
}

// SerializeIndex saves the index to a file
func (f *S2Finder) SerializeIndex(filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(f)
	return err
}

// DeserializeIndex loads the index from a file
func DeserializeIndex(filepath string) (*S2Finder, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	var finder S2Finder
	err = decoder.Decode(&finder)
	if err != nil {
		return nil, fmt.Errorf("error decoding file: %v", err)
	}

	return &finder, nil
}

func init() {
	gob.Register(&s2.RegionCoverer{})
	gob.Register(s2.CellID(0))
	gob.Register(SerializableSpatialCity{})
	gob.Register(SerializableRect{})
}
