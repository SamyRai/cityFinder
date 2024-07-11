// cmd/server/main_test.go
package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/SamyRai/cityFinder/city"
	"github.com/SamyRai/cityFinder/finder"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Mock data for testing
var mockData = loadRealIndex()

func loadRealIndex() *finder.S2Finder {
	s2indexPath := "./../../datasets/s2index.gob"
	fmt.Printf("Loading S2 index from %s", s2indexPath)

	s2Finder, err := finder.DeserializeIndex(s2indexPath)
	if err != nil {
		panic(err)
	}

	return s2Finder
}

func setupMockApp() *fiber.App {
	app := fiber.New()
	setupRoutes(app, mockData)
	return app
}

func pickRandomLines(filepath string, count int) ([]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))
	rand.Shuffle(len(lines), func(i, j int) {
		lines[i], lines[j] = lines[j], lines[i]
	})

	if count > len(lines) {
		count = len(lines)
	}

	return lines[:count], nil
}

func parseCityLine(line string) (city.City, error) {
	fields := strings.Split(line, "\t")
	if len(fields) < 9 {
		return city.City{}, fmt.Errorf("invalid line format")
	}

	lat, err := strconv.ParseFloat(fields[4], 64)
	if err != nil {
		return city.City{}, err
	}

	lon, err := strconv.ParseFloat(fields[5], 64)
	if err != nil {
		return city.City{}, err
	}

	return city.City{
		Latitude:  lat,
		Longitude: lon,
		Name:      fields[1],
		Country:   fields[8],
	}, nil
}

func pickRandomPostalCodes(filepath string, count int) ([]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '\t'
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(records), func(i, j int) {
		records[i], records[j] = records[j], records[i]
	})

	var postalCodes []string
	for i := 0; i < count && i < len(records); i++ {
		postalCodes = append(postalCodes, records[i][1])
	}

	return postalCodes, nil
}

func TestGetNearestCityRandom(t *testing.T) {
	app := setupMockApp()
	lines, err := pickRandomLines("./../../datasets/allCountries.csv", 20)
	if err != nil {
		t.Fatalf("Failed to pick random lines: %v", err)
	}

	for _, line := range lines {
		loc, err := parseCityLine(line)
		if err != nil {
			t.Fatalf("Failed to parse city line: %v", err)
		}

		t.Run(loc.Name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/nearest?lat="+fmt.Sprintf("%f", loc.Latitude)+"&lon="+fmt.Sprintf("%f", loc.Longitude), nil)
			resp, _ := app.Test(req, -1)

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var cityObj city.City
			err = json.NewDecoder(resp.Body).Decode(&cityObj)
			assert.NoError(t, err)
			assert.NotEmpty(t, cityObj.Name)
		})
	}
}

func TestGetCoordinatesByNameRandom(t *testing.T) {
	app := setupMockApp()
	lines, err := pickRandomLines("./../../datasets/allCountries.csv", 20)
	if err != nil {
		t.Fatalf("Failed to pick random lines: %v", err)
	}

	for _, line := range lines {
		loc, err := parseCityLine(line)
		if err != nil {
			t.Fatalf("Failed to parse city line: %v", err)
		}

		t.Run(loc.Name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/coordinates?name="+url.QueryEscape(loc.Name), nil)
			resp, _ := app.Test(req, -1)

			if resp.StatusCode != http.StatusOK {
				t.Logf("City %s not found, skipping...", loc.Name)
				return
			}

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var cityObj city.City
			err = json.NewDecoder(resp.Body).Decode(&cityObj)
			assert.NoError(t, err)
			assert.Equal(t, loc.Latitude, cityObj.Latitude)
			assert.Equal(t, loc.Longitude, cityObj.Longitude)
		})
	}
}

func TestGetCityByPostalCodeRandom(t *testing.T) {
	app := setupMockApp()
	postalCodes, err := pickRandomPostalCodes("./../../datasets/zipCodes.csv", 20)
	if err != nil {
		t.Fatalf("Failed to pick random postal codes: %v", err)
	}

	for _, code := range postalCodes {
		t.Run(code, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/postalcode?postalcode="+url.QueryEscape(code), nil)
			resp, _ := app.Test(req, -1)

			if resp.StatusCode != http.StatusOK {
				t.Logf("Postal code %s not found, skipping...", code)
				return
			}
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var cityObj city.City
			err = json.NewDecoder(resp.Body).Decode(&cityObj)
			assert.NoError(t, err)
			assert.NotEmpty(t, cityObj.Name)
		})
	}
}

func TestBadRequest(t *testing.T) {
	app := setupMockApp()

	req := httptest.NewRequest("GET", "/nearest?lat=invalid&lon=-74.0060", nil)
	resp, _ := app.Test(req, -1)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	bodyString := string(bodyBytes)

	assert.Equal(t, "Invalid latitude", bodyString)
}
