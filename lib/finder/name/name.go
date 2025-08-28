package name

import (
	"encoding/gob"
	"github.com/SamyRai/cityFinder/lib/city"
	"github.com/SamyRai/cityFinder/util"
	"github.com/cheggaaa/pb/v3"
	"os"
	"sync"
)

// Finder is a struct that contains the data for city name lookups
type Finder struct {
	InvertedIndex map[string]map[string][]*city.City // Inverted index for city name lookups by country
	BKTree        *util.BKTree                       // BK-tree for fuzzy city name matching
	mutex         sync.RWMutex                       // Mutex for thread-safe operations
}

// NewNameFinder creates a new NameFinder instance
func NewNameFinder() *Finder {
	return &Finder{
		InvertedIndex: make(map[string]map[string][]*city.City),
		BKTree:        util.NewBKTree(),
	}
}

// BuildIndex creates a name index from city data
func BuildIndex(cities []city.SpatialCity) *Finder {
	finder := NewNameFinder()
	bar := pb.Full.Start(len(cities))
	for _, spatialCity := range cities {
		finder.AddCity(spatialCity)
		bar.Increment()
	}
	bar.Finish()
	return finder
}

// AddCity adds a city to the NameFinder
func (nf *Finder) AddCity(spatialCity city.SpatialCity) {
	names := append(spatialCity.AltNames, spatialCity.Name)
	for _, name := range names {
		nf.mutex.Lock()
		if _, exists := nf.InvertedIndex[spatialCity.Country]; !exists {
			nf.InvertedIndex[spatialCity.Country] = make(map[string][]*city.City)
		}
		nf.InvertedIndex[spatialCity.Country][name] = append(nf.InvertedIndex[spatialCity.Country][name], &spatialCity.City)
		nf.BKTree.Add(name)
		nf.mutex.Unlock()
	}
}

// CityByName finds the coordinates of a city by its name
func (nf *Finder) CityByName(name string, countryCode string) *city.City {
	nf.mutex.RLock()
	defer nf.mutex.RUnlock()

	if cities, exists := nf.InvertedIndex[countryCode][name]; exists && len(cities) > 0 {
		return cities[0] // Return the first match if an exact match is found
	}

	// Perform fuzzy search using BK-tree if no exact match is found
	candidates := nf.BKTree.Search(name, 2) // Adjust the distance threshold as needed
	if len(candidates) > 0 {
		for _, candidate := range candidates {
			if cities, exists := nf.InvertedIndex[countryCode][candidate]; exists && len(cities) > 0 {
				return cities[0]
			}
		}
	}

	return nil
}

// SerializeIndex saves the name index to a file
func (nf *Finder) SerializeIndex(filepath string) error {
	nf.mutex.Lock()
	defer nf.mutex.Unlock()

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(nf)
	return err
}

// DeserializeIndex loads the name index from a file
func DeserializeIndex(filepath string) (*Finder, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	var finder Finder
	err = decoder.Decode(&finder)
	if err != nil {
		return nil, err
	}

	return &finder, nil
}
