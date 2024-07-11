package main

import (
	"cityFinder/city"
	"cityFinder/dataloader"
	"cityFinder/finder"
	"fmt"
	"log"
	"sync"
	"time"
)

type Result struct {
	Label       string
	Duration    time.Duration
	NearestCity *city.City
}

func MeasureTime(wg *sync.WaitGroup, resultsChan chan<- Result, label string, f func() *city.City) {
	defer wg.Done()
	start := time.Now()
	nearestCity := f()
	duration := time.Since(start)
	resultsChan <- Result{Label: label, Duration: duration, NearestCity: nearestCity}
}

var testLocations = []struct {
	Lat      float64
	Lon      float64
	Expected string
}{
	{40.7128, -74.0060, "New York"},
	{34.0522, -118.2437, "Los Angeles"},
	{41.8781, -87.6298, "Chicago"},
	{51.5074, -0.1278, "London"},
	{48.8566, 2.3522, "Paris"},
	{35.6895, 139.6917, "Tokyo"},
	{55.7558, 37.6176, "Moscow"},
	{-33.8688, 151.2093, "Sydney"},
	{39.9042, 116.4074, "Beijing"},
	{19.4326, -99.1332, "Mexico City"},
}

func main() {
	var cities []city.SpatialCity
	var err error
	csvLocation := "datasets/allCountries.csv"
	postalCodeLocation := "datasets/zipCodes.csv"

	var wg sync.WaitGroup
	resultsChan := make(chan Result, len(testLocations)*5)

	log.Printf("Loading the CSV data from %s", csvLocation)
	cities, err = dataloader.LoadGeoNamesCSV(csvLocation)
	if err != nil {
		log.Fatalf("Failed to load GeoNames data from CSV: %v", err)
	}

	log.Printf("Loading the Postal Code data from %s", postalCodeLocation)
	postalCodes, err := dataloader.LoadPostalCodes(postalCodeLocation)
	if err != nil {
		log.Fatalf("Failed to load Postal Code data: %v", err)
	}

	// Initialize all finders
	finders := map[string]finder.Finder{
		"R-tree":   finder.BuildRTree(cities, postalCodes),
		"Geohash":  finder.BuildGeoHashIndex(cities, 12, postalCodes),
		"S2":       finder.BuildS2Index(cities, postalCodes),
		"k-d Tree": finder.BuildKDTree(cities, postalCodes),
	}

	for name, f := range finders {
		for _, loc := range testLocations {
			wg.Add(1)
			go MeasureTime(&wg, resultsChan, fmt.Sprintf("Finding nearest city using %s for %s", name, loc.Expected), func() *city.City {
				return f.FindNearestCity(loc.Lat, loc.Lon)
			})
		}
	}

	// Example postal code search
	postalCodesToTest := []string{"10001", "90210", "60601"} // Example postal codes
	for name, f := range finders {
		for _, postalCode := range postalCodesToTest {
			wg.Add(1)
			go MeasureTime(&wg, resultsChan, fmt.Sprintf("Finding nearest city using %s for postal code %s", name, postalCode), func() *city.City {
				return f.FindCityByPostalCode(postalCode)
			})
		}
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	results := []Result{}
	for result := range resultsChan {
		results = append(results, result)
	}

	fmt.Println("\nSummary of Results:")
	for _, result := range results {
		if result.NearestCity != nil {
			fmt.Printf("%s: took %v, Nearest city: %s, %s\n", result.Label, result.Duration, result.NearestCity.Name, result.NearestCity.Country)
		} else {
			fmt.Printf("%s: took %v\n", result.Label, result.Duration)
		}
	}
}
