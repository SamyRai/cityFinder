package main

import (
	"encoding/csv"
	"fmt"
	"github.com/SamyRai/cityFinder/benchmark"
	"github.com/SamyRai/cityFinder/dataloader"
	"github.com/SamyRai/cityFinder/finder"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

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
	{55.7963, 49.1088, "Kazan"},
	{54.5378, 52.7985, "Bugulma"},
}

var testPostalCodes = []string{
	"10001", "90210", "60601",
}

func main() {
	csvLocation := "datasets/allCountries.csv"
	postalCodeLocation := "datasets/zipCodes.csv"

	log.Printf("Loading the CSV data from %s", csvLocation)
	cities, err := dataloader.LoadGeoNamesCSV(csvLocation)
	if err != nil {
		log.Fatalf("Failed to load GeoNames data from CSV: %v", err)
	}
	log.Printf("Finished loading CSV data")

	log.Printf("Loading the Postal Code data from %s", postalCodeLocation)
	postalCodes, err := dataloader.LoadPostalCodes(postalCodeLocation)
	if err != nil {
		log.Fatalf("Failed to load Postal Code data: %v", err)
	}
	log.Printf("Finished loading Postal Code data")

	// Clean up the datasets after loading
	defer func() {
		cities = nil
		postalCodes = nil
		runtime.GC()
	}()

	// Initialize all finders and measure their memory consumption
	log.Printf("Initializing finders")
	var memStatsBefore, memStatsAfter runtime.MemStats

	finders := make(map[string]finder.Finder)
	overallMemoryUsage := make(map[string]uint64)
	var wg sync.WaitGroup
	var mu sync.Mutex

	initFinder := func(name string, buildFunc func() finder.Finder) {
		defer wg.Done()
		log.Printf("Initializing %s finder", name)
		runtime.ReadMemStats(&memStatsBefore)
		f := buildFunc()
		runtime.ReadMemStats(&memStatsAfter)
		mu.Lock()
		finders[name] = f
		overallMemoryUsage[name] = memStatsAfter.Alloc - memStatsBefore.Alloc
		mu.Unlock()
		log.Printf("Finished initializing %s finder", name)
	}

	wg.Add(3)
	go initFinder("R-tree", func() finder.Finder { return finder.BuildRTree(cities, postalCodes) })
	go initFinder("Geohash", func() finder.Finder { return finder.BuildGeoHashIndex(cities, 12, postalCodes) })
	go initFinder("k-d Tree", func() finder.Finder { return finder.BuildKDTree(cities, postalCodes) })
	wg.Wait()

	log.Printf("Finished initializing all finders")

	log.Printf("Running benchmarks")
	start := time.Now()
	results := benchmark.BenchmarkFinders(finders, overallMemoryUsage, testLocations, testPostalCodes)
	duration := time.Since(start)
	log.Printf("Finished running benchmarks in %v", duration)

	fmt.Println("\nSummary of Results:")
	printResultsTable(results, overallMemoryUsage)

	log.Printf("Saving results to CSV")
	saveResultsToCSV(results, overallMemoryUsage, "results.csv")
	log.Printf("Finished saving results to CSV")
}

func printResultsTable(results []benchmark.Result, overallMemoryUsage map[string]uint64) {
	cityResults := make(map[string][]benchmark.Result)
	for _, result := range results {
		city := extractCityName(result.Label)
		cityResults[city] = append(cityResults[city], result)
	}

	header := fmt.Sprintf("%-40s %-20s %-15s %-15s %-20s %-15s %-15s", "City", "Finder", "Time (ms)", "Memory (KB)", "Nearest City", "Latitude", "Longitude")
	fmt.Println(header)
	fmt.Println(strings.Repeat("-", len(header)))

	for city, results := range cityResults {
		var fastestResult *benchmark.Result
		var lowestMemoryResult *benchmark.Result
		for i, result := range results {
			if fastestResult == nil || result.Duration < fastestResult.Duration {
				fastestResult = &results[i]
			}
			if lowestMemoryResult == nil || result.MemoryUsage < lowestMemoryResult.MemoryUsage {
				lowestMemoryResult = &results[i]
			}
		}

		for _, result := range results {
			isFastest := result == *fastestResult
			isLowestMemory := result == *lowestMemoryResult
			finderName := extractFinderName(result.Label)
			time := result.Duration.Milliseconds()
			memory := result.MemoryUsage / 1024
			nearestCityName := "N/A"
			latitude := "N/A"
			longitude := "N/A"
			if result.NearestCity != nil {
				nearestCityName = result.NearestCity.Name
				latitude = fmt.Sprintf("%f", result.NearestCity.Latitude)
				longitude = fmt.Sprintf("%f", result.NearestCity.Longitude)
			}

			timeStr := fmt.Sprintf("%d", time)
			if isFastest {
				timeStr += " <- Fastest"
			}

			memoryStr := fmt.Sprintf("%d", memory)
			if isLowestMemory {
				memoryStr += " <- Lowest"
			}

			fmt.Printf("%-40s %-20s %-15s %-15s %-20s %-15s %-15s\n", city, finderName, timeStr, memoryStr, nearestCityName, latitude, longitude)
		}
	}

	fmt.Println("\nOverall Memory Consumption:")
	header = fmt.Sprintf("%-20s %-15s", "Finder", "Memory (KB)")
	fmt.Println(header)
	fmt.Println(strings.Repeat("-", len(header)))

	for finderName, memory := range overallMemoryUsage {
		fmt.Printf("%-20s %-15d\n", finderName, memory/1024)
	}
}

func saveResultsToCSV(results []benchmark.Result, overallMemoryUsage map[string]uint64, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"City", "Finder", "Time", "Memory", "Nearest City", "Latitude", "Longitude", "Expected Latitude", "Expected Longitude"}
	writer.Write(header)

	cityResults := make(map[string][]benchmark.Result)
	for _, result := range results {
		city := extractCityName(result.Label)
		cityResults[city] = append(cityResults[city], result)
	}

	for city, results := range cityResults {
		expectedLat, expectedLon := getExpectedCoordinates(city)
		for _, result := range results {
			finderName := extractFinderName(result.Label)
			time := result.Duration.Milliseconds()
			memory := result.MemoryUsage / 1024
			nearestCityName := "N/A"
			latitude := "N/A"
			longitude := "N/A"
			if result.NearestCity != nil {
				nearestCityName = result.NearestCity.Name
				latitude = fmt.Sprintf("%f", result.NearestCity.Latitude)
				longitude = fmt.Sprintf("%f", result.NearestCity.Longitude)
			}

			record := []string{city, finderName, fmt.Sprintf("%d", time), fmt.Sprintf("%d", memory), nearestCityName, latitude, longitude, fmt.Sprintf("%f", expectedLat), fmt.Sprintf("%f", expectedLon)}
			writer.Write(record)
		}
	}
}

func extractCityName(label string) string {
	parts := strings.Split(label, " for ")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

func extractFinderName(label string) string {
	parts := strings.Split(label, " using ")
	if len(parts) == 2 {
		return strings.Split(parts[1], " for ")[0]
	}
	return ""
}

func getExpectedCoordinates(cityName string) (float64, float64) {
	for _, loc := range testLocations {
		if loc.Expected == cityName {
			return loc.Lat, loc.Lon
		}
	}
	return 0.0, 0.0
}
