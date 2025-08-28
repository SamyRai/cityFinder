package postalCode

import (
	"encoding/gob"
	"github.com/SamyRai/cityFinder/lib/city"
	"github.com/SamyRai/cityFinder/lib/dataLoader"
	"github.com/cheggaaa/pb/v3"
	"os"
	"sync"
)

// Finder is a struct that contains the data for postal code lookups
type Finder struct {
	PostalCode map[string]map[string]dataLoader.PostalCodeEntry // Map of country code to postal code to entry
	mutex      sync.RWMutex                                     // Mutex for thread-safe operations
}

// NewPostalCodeFinder creates a new Finder instance
func NewPostalCodeFinder() *Finder {
	return &Finder{
		PostalCode: make(map[string]map[string]dataLoader.PostalCodeEntry),
	}
}

// AddPostalCode adds a postal code entry to the Finder
func (pcf *Finder) AddPostalCode(entry dataLoader.PostalCodeEntry) {
	pcf.mutex.Lock()
	defer pcf.mutex.Unlock()

	if _, exists := pcf.PostalCode[entry.CountryCode]; !exists {
		pcf.PostalCode[entry.CountryCode] = make(map[string]dataLoader.PostalCodeEntry)
	}
	pcf.PostalCode[entry.CountryCode][entry.PostalCode] = entry
}

// BuildIndex creates a postal code index from postal code data
func BuildIndex(postalCodes map[string]map[string]dataLoader.PostalCodeEntry) *Finder {
	finder := NewPostalCodeFinder()
	total := 0
	for _, countryCode := range postalCodes {
		for range countryCode {
			total++
		}
	}
	bar := pb.Full.Start(total)
	for _, countryCode := range postalCodes {
		for _, entry := range countryCode {
			finder.AddPostalCode(entry)
			bar.Increment()
		}
	}
	bar.Finish()

	return finder
}

// CityByPostalCode finds the nearest city by postal code and country code
func (pcf *Finder) CityByPostalCode(postalCode, countryCode string) *city.City {
	pcf.mutex.RLock()
	defer pcf.mutex.RUnlock()

	if countryEntries, exists := pcf.PostalCode[countryCode]; exists {
		if entry, exists := countryEntries[postalCode]; exists {
			return &city.City{
				Latitude:  entry.Latitude,
				Longitude: entry.Longitude,
				Name:      entry.PlaceName,
				Country:   countryCode,
			}
		}
	}
	return nil
}

// SerializeIndex saves the postal code index to a file
func (pcf *Finder) SerializeIndex(filepath string) error {
	pcf.mutex.Lock()
	defer pcf.mutex.Unlock()

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(pcf)
	return err
}

// DeserializeIndex loads the postal code index from a file
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
