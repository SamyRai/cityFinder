package dataLoader

import (
	"bufio"
	"fmt"
	"log"
	"github.com/SamyRai/cityFinder/lib/city"
	"os"
	"strconv"
	"strings"
)

func LoadGeoNamesCSV(filepath string) ([]city.SpatialCity, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var cities []city.SpatialCity
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("Processing line: %s", line)
		fields := strings.Split(line, "\t")
		if len(fields) < 19 {
			log.Printf("Skipping line with %d fields: %s", len(fields), line)
			continue
		}

		lat, err := strconv.ParseFloat(fields[4], 64)
		if err != nil {
			log.Printf("Error parsing lat: %v on line: %s\n", err, line)
			continue
		}
		lon, err := strconv.ParseFloat(fields[5], 64)
		if err != nil {
			log.Printf("Error parsing lon: %v on line: %s\n", err, line)
			continue
		}

		altNames := strings.Split(fields[3], ",")

		cityObj := city.City{
			Latitude:  lat,
			Longitude: lon,
			Name:      fields[1],
			Country:   fields[8],
			AltNames:  altNames,
		}

		rect := &city.Rect{
			Min: []float64{lon - 0.00001, lat - 0.00001},
			Max: []float64{lon + 0.00001, lat + 0.00001},
		}
		spatialCity := city.SpatialCity{City: cityObj, Rect: rect}

		cities = append(cities, spatialCity)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan file: %v, %v", filepath, err)
	}
	log.Printf("Loaded %d cities from %s\n", len(cities), filepath)
	return cities, nil
}

func StreamGeoNamesCSV(filepath string, cityChan chan<- city.SpatialCity, errChan chan<- error) {
	file, err := os.Open(filepath)
	if err != nil {
		errChan <- err
		close(cityChan)
		close(errChan)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, "\t")
		if len(fields) < 9 {
			continue
		}

		lat, err := strconv.ParseFloat(fields[4], 64)
		if err != nil {
			continue
		}
		lon, err := strconv.ParseFloat(fields[5], 64)
		if err != nil {
			continue
		}

		cityObj := city.City{
			Latitude:  lat,
			Longitude: lon,
			Name:      fields[1],
			Country:   fields[8],
		}

		rect := &city.Rect{
			Min: []float64{lon - 0.00001, lat - 0.00001},
			Max: []float64{lon + 0.00001, lat + 0.00001},
		}
		spatialCity := city.SpatialCity{City: cityObj, Rect: rect}

		cityChan <- spatialCity
	}

	if err := scanner.Err(); err != nil {
		errChan <- err
	}

	close(cityChan)
	close(errChan)
}
