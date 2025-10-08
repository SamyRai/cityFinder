package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/SamyRai/cityFinder/benchmark"
	"github.com/SamyRai/cityFinder/lib/config"
	"github.com/SamyRai/cityFinder/lib/finder"
	"github.com/SamyRai/cityFinder/lib/initializer"
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

func main() {
	log.Println("Benchmark started.")
	configPath, exists := os.LookupEnv("CONFIG_PATH")
	if !exists {
		configPath = "config.json"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	mainFinder, err := initializer.Initialize(cfg)
	if err != nil {
		log.Fatalf("Initialization failed: %v", err)
	}

	// Clean up the datasets after loading
	defer func() {
		runtime.GC()
	}()

	finders := map[string]*finder.Finder{
		"S2": mainFinder,
	}
	overallMemoryUsage := make(map[string]uint64)

	log.Printf("Running benchmarks")
	start := time.Now()
	results := benchmark.BenchmarkFinders(finders, overallMemoryUsage, testLocations)
	duration := time.Since(start)
	log.Printf("Finished running benchmarks in %v", duration)

	fmt.Println("\nSummary of Results:")
	printResultsTable(results, overallMemoryUsage)

	log.Printf("Saving results to CSV")
	saveResultsToCSV(results, "results.csv")
	log.Printf("Finished saving results to CSV")
}

func printResultsTable(results []benchmark.Result, overallMemoryUsage map[string]uint64) {
	cityResults := make(map[string][]benchmark.Result)
	for _, result := range results {
		city := extractCityName(result.Label)
		cityResults[city] = append(cityResults[city], result)
	}

	header := fmt.Sprintf("%-40s %-20s %-15s %-15s %-20s %-15s %-15s", "City", "Finder", "Time (ns)", "Memory (B)", "Nearest City", "Latitude", "Longitude")
	fmt.Println(header)
	fmt.Println(strings.Repeat("-", len(header)))

	for city, cityRes := range cityResults {
		for _, result := range cityRes {
			finderName := extractFinderName(result.Label)
			time := result.Duration.Nanoseconds()
			memory := result.MemoryUsage
			nearestCityName := "N/A"
			latitude := "N/A"
			longitude := "N/A"
			if result.NearestCity != nil {
				nearestCityName = result.NearestCity.Name
				latitude = fmt.Sprintf("%f", result.NearestCity.Latitude)
				longitude = fmt.Sprintf("%f", result.NearestCity.Longitude)
			}

			fmt.Printf("%-40s %-20s %-15d %-15d %-20s %-15s %-15s\n", city, finderName, time, memory, nearestCityName, latitude, longitude)
		}
	}
}

func saveResultsToCSV(results []benchmark.Result, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create CSV file: %v", err)
	}
	writer := csv.NewWriter(file)

	header := []string{"City", "Finder", "Time (ns)", "Memory (B)", "Nearest City", "Latitude", "Longitude"}
	if err := writer.Write(header); err != nil {
		log.Printf("Warning: failed to write CSV header: %v", err)
	}

	for _, result := range results {
		city := extractCityName(result.Label)
		finderName := extractFinderName(result.Label)
		time := result.Duration.Nanoseconds()
		memory := result.MemoryUsage
		nearestCityName := "N/A"
		latitude := "N/A"
		longitude := "N/A"
		if result.NearestCity != nil {
			nearestCityName = result.NearestCity.Name
			latitude = fmt.Sprintf("%f", result.NearestCity.Latitude)
			longitude = fmt.Sprintf("%f", result.NearestCity.Longitude)
		}

		record := []string{city, finderName, fmt.Sprintf("%d", time), fmt.Sprintf("%d", memory), nearestCityName, latitude, longitude}
		if err := writer.Write(record); err != nil {
			log.Printf("Warning: failed to write CSV record: %v", err)
		}
	}
	writer.Flush()
	if err := file.Close(); err != nil {
		log.Printf("Warning: failed to close CSV file: %v", err)
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