package benchmark

import (
	"fmt"
	"github.com/SamyRai/cityFinder/city"
	"github.com/SamyRai/cityFinder/finder"
	"runtime"
	"sync"
	"time"
)

type Result struct {
	Label       string
	Duration    time.Duration
	MemoryUsage uint64
	NearestCity *city.City
}

func MeasureTimeAndMemory(wg *sync.WaitGroup, resultsChan chan<- Result, label string, f func() *city.City) {
	defer wg.Done()
	var memStatsBefore, memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)
	start := time.Now()
	nearestCity := f()
	duration := time.Since(start)
	runtime.ReadMemStats(&memStatsAfter)

	memoryUsage := memStatsAfter.Alloc - memStatsBefore.Alloc
	resultsChan <- Result{Label: label, Duration: duration, MemoryUsage: memoryUsage, NearestCity: nearestCity}
}

func BenchmarkFinders(finders map[string]finder.Finder, overallMemoryUsage map[string]uint64, testLocations []struct {
	Lat, Lon float64
	Expected string
}, postalCodes []string) []Result {
	var wg sync.WaitGroup
	resultsChan := make(chan Result, len(testLocations)*len(finders)+len(postalCodes)*len(finders))

	for name, f := range finders {
		for _, loc := range testLocations {
			wg.Add(1)
			go MeasureTimeAndMemory(&wg, resultsChan, fmt.Sprintf("Finding nearest city using %s for %s", name, loc.Expected), func() *city.City {
				return f.FindNearestCity(loc.Lat, loc.Lon)
			})
		}
		for _, postalCode := range postalCodes {
			wg.Add(1)
			go MeasureTimeAndMemory(&wg, resultsChan, fmt.Sprintf("Finding nearest city using %s for postal code %s", name, postalCode), func() *city.City {
				return f.FindCityByPostalCode(postalCode)
			})
		}
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var results []Result
	for result := range resultsChan {
		results = append(results, result)
	}

	return results
}
