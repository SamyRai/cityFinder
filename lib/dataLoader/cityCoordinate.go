package dataLoader

import (
	"bufio"
	"fmt"
	"github.com/SamyRai/cityFinder/lib/city"
	"os"
	"strconv"
	"strings"

	"github.com/cheggaaa/pb/v3"
	"github.com/dhconnelly/rtreego"
)

func LoadGeoNamesCSV(filepath string) ([]city.SpatialCity, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Count the number of lines in the file for the progress bar
	lineCount := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineCount++
	}
	// Reset the file pointer to the beginning
	_, err = file.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to seek to beginning of file: %v", err)
	}

	var cities []city.SpatialCity
	scanner = bufio.NewScanner(file)
	bar := pb.Full.Start(lineCount)
	defer bar.Finish()

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

		altNames := strings.Split(fields[3], ",")

		cityObj := city.City{
			Latitude:  lat,
			Longitude: lon,
			Name:      fields[1],
			Country:   fields[8],
			AltNames:  altNames,
		}

		point := rtreego.Point{lon, lat}
		rect, _ := rtreego.NewRect(point, []float64{0.00001, 0.00001})
		spatialCity := city.SpatialCity{City: cityObj, Rect: rect}

		cities = append(cities, spatialCity)
		bar.Increment()
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan file: %v, %v", filepath, err)
	}
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

		point := rtreego.Point{lon, lat}
		rect, _ := rtreego.NewRect(point, []float64{0.00001, 0.00001})
		spatialCity := city.SpatialCity{City: cityObj, Rect: rect}

		cityChan <- spatialCity
	}

	if err := scanner.Err(); err != nil {
		errChan <- err
	}

	close(cityChan)
	close(errChan)
}
