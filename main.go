package main

import (
	"cityFinder/binary"
	"cityFinder/city"
	"cityFinder/csv"
	"cityFinder/rtree"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/dhconnelly/rtreego"
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
	var rtreeObj, loadedRTree *rtreego.Rtree
	var err error
	binaryLocation := "cities.bin"
	rtreeLocation := "rtree.gob"
	textLocation := "allCountries.txt"

	var wg sync.WaitGroup
	var rtreeMutex sync.Mutex
	resultsChan := make(chan Result, len(testLocations)*3+6)

	log.Printf("Loading the CSV data from %s", textLocation)
	wg.Add(1)
	go MeasureTime(&wg, resultsChan, "Loading and building R-tree from CSV", func() *city.City {
		cities, err = csv.LoadGeoNamesCSV(textLocation)
		if err != nil {
			log.Fatalf("Failed to load GeoNames data from CSV: %v", err)
		}
		rtreeMutex.Lock()
		rtreeObj = rtree.BuildRTree(cities)
		rtreeMutex.Unlock()
		return nil
	})

	for _, loc := range testLocations {
		wg.Add(1)
		go MeasureTime(&wg, resultsChan, fmt.Sprintf("Finding nearest city using CSV data for %s", loc.Expected), func() *city.City {
			rtreeMutex.Lock()
			defer rtreeMutex.Unlock()
			if rtreeObj != nil {
				return rtree.FindNearestCity(loc.Lat, loc.Lon, rtreeObj)
			}
			return nil
		})
	}

	log.Printf("Saving the binary data to %s", binaryLocation)
	wg.Add(1)
	go MeasureTime(&wg, resultsChan, "Saving the binary from text", func() *city.City {
		err := binary.SaveBinary(binaryLocation, cities)
		if err != nil {
			log.Fatalf("Failed to save binary data: %v", err)
		}
		return nil
	})

	log.Printf("Loading the binary data from %s", binaryLocation)
	wg.Add(1)
	go MeasureTime(&wg, resultsChan, "Loading and building R-tree from binary", func() *city.City {
		cities, err = binary.LoadBinary(binaryLocation)
		if err != nil {
			log.Fatalf("Failed to load binary data: %v", err)
		}
		rtreeMutex.Lock()
		rtreeObj = rtree.BuildRTree(cities)
		rtreeMutex.Unlock()
		return nil
	})

	for _, loc := range testLocations {
		wg.Add(1)
		go MeasureTime(&wg, resultsChan, fmt.Sprintf("Finding nearest city using binary data for %s", loc.Expected), func() *city.City {
			rtreeMutex.Lock()
			defer rtreeMutex.Unlock()
			if rtreeObj != nil {
				return rtree.FindNearestCity(loc.Lat, loc.Lon, rtreeObj)
			}
			return nil
		})
	}

	if _, err := os.Stat(rtreeLocation); os.IsNotExist(err) {
		log.Printf("Saving the R-tree data to %s", rtreeLocation)
		wg.Add(1)
		go MeasureTime(&wg, resultsChan, "Saving the R-tree", func() *city.City {
			rtreeMutex.Lock()
			defer rtreeMutex.Unlock()
			err := rtree.SaveRTree(rtreeLocation, rtreeObj)
			if err != nil {
				log.Fatalf("Failed to save R-tree data: %v", err)
			}
			return nil
		})
	} else {
		log.Printf("Loading the R-tree data from %s", rtreeLocation)
		wg.Add(1)
		go MeasureTime(&wg, resultsChan, "Loading the R-tree", func() *city.City {
			loadedRTree, err = rtree.LoadRTree(rtreeLocation)
			if err != nil {
				log.Fatalf("Failed to load R-tree data: %v", err)
			}
			return nil
		})

		for _, loc := range testLocations {
			wg.Add(1)
			go MeasureTime(&wg, resultsChan, fmt.Sprintf("Finding nearest city using pre-built R-tree data for %s", loc.Expected), func() *city.City {
				if loadedRTree != nil {
					return rtree.FindNearestCity(loc.Lat, loc.Lon, loadedRTree)
				}
				return nil
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
