package csv

import (
	"bufio"
	"cityFinder/city"
	"os"
	"strconv"
	"strings"

	"github.com/cheggaaa/pb/v3"
	"github.com/dhconnelly/rtreego"
)

func LoadGeoNamesCSV(filepath string) ([]city.SpatialCity, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	lineCount := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineCount++
	}
	file.Seek(0, 0)

	var cities []city.SpatialCity
	scanner = bufio.NewScanner(file)
	currentLine := 0
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

		cityObj := city.City{
			Latitude:  lat,
			Longitude: lon,
			Name:      fields[1],
			Country:   fields[8],
		}

		point := rtreego.Point{lon, lat}
		rect, _ := rtreego.NewRect(point, []float64{0.00001, 0.00001})
		spatialCity := city.SpatialCity{City: cityObj, Rect: rect}

		cities = append(cities, spatialCity)

		currentLine++
		bar.Increment()
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return cities, nil
}
