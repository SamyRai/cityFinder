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
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

type ServerTestSuite struct {
	suite.Suite
	app    *fiber.App
	finder finder.Finder
}

func (suite *ServerTestSuite) SetupSuite() {
	// For now, we only test the S2 finder.
	finderType := "s2"
	csv := "./../../datasets/allCountries_small.csv"
	cities := fmt.Sprintf("./../../datasets/cities_%s_test.gob", finderType)
	points := fmt.Sprintf("./../../datasets/points_%s_test.gob", finderType)
	meta := fmt.Sprintf("./../../datasets/meta_%s_test.gob", finderType)

	cmd := exec.Command("go", "run", "./../../cmd/builder",
		"-finder="+finderType,
		"-csv="+csv,
		"-cities="+cities,
		"-points="+points,
		"-meta="+meta)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to build test index: %v\nOutput:\n%s", err, output)
	}

	suite.finder, err = finder.DeserializeS2(meta, cities, points)
	require.NoError(suite.T(), err)

	suite.app = setupMockAppTestify(suite.finder)
}

func (suite *ServerTestSuite) TearDownSuite() {
	suite.finder.Close()
	finderType := "s2"
	os.Remove(fmt.Sprintf("./../../datasets/cities_%s_test.gob", finderType))
	os.Remove(fmt.Sprintf("./../../datasets/points_%s_test.gob", finderType))
	os.Remove(fmt.Sprintf("./../../datasets/meta_%s_test.gob", finderType))
}

func setupMockAppTestify(f finder.Finder) *fiber.App {
	app := fiber.New()
	setupRoutes(app, f)
	return app
}

func (suite *ServerTestSuite) pickRandomLines(filepath string, count int) ([]string, error) {
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

func (suite *ServerTestSuite) parseCityLine(line string) (city.City, error) {
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

func (suite *ServerTestSuite) pickRandomPostalCodes(filepath string, count int) ([]string, error) {
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

func (suite *ServerTestSuite) TestGetNearestCityRandom() {
	suite.T().Skip("Skipping test due to OOM issues")
	lines, err := suite.pickRandomLines("./../../datasets/allCountries_small.csv", 20)
	require.NoError(suite.T(), err)

	for _, line := range lines {
		loc, err := suite.parseCityLine(line)
		require.NoError(suite.T(), err)

		suite.Run(loc.Name, func() {
			req := httptest.NewRequest("GET", "/nearest?lat="+fmt.Sprintf("%f", loc.Latitude)+"&lon="+fmt.Sprintf("%f", loc.Longitude), nil)
			resp, _ := suite.app.Test(req, -1)

			assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

			var cityObj city.City
			err := json.NewDecoder(resp.Body).Decode(&cityObj)
			assert.NoError(suite.T(), err)
			assert.NotEmpty(suite.T(), cityObj.Name)
		})
	}
}

func (suite *ServerTestSuite) TestGetCoordinatesByNameRandom() {
	suite.T().Skip("Skipping test due to OOM issues")
	lines, err := suite.pickRandomLines("./../../datasets/allCountries_small.csv", 20)
	require.NoError(suite.T(), err)

	for _, line := range lines {
		loc, err := suite.parseCityLine(line)
		require.NoError(suite.T(), err)

		suite.Run(loc.Name, func() {
			req := httptest.NewRequest("GET", "/coordinates?name="+url.QueryEscape(loc.Name), nil)
			resp, _ := suite.app.Test(req, -1)

			if resp.StatusCode != http.StatusOK {
				suite.T().Logf("City %s not found, skipping...", loc.Name)
				return
			}

			assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

			var cities []*city.City
			err := json.NewDecoder(resp.Body).Decode(&cities)
			assert.NoError(suite.T(), err)
			assert.NotEmpty(suite.T(), cities)
		})
	}
}

func (suite *ServerTestSuite) TestGetCityByPostalCodeRandom() {
	suite.T().Skip("Skipping test due to OOM issues")
	postalCodes, err := suite.pickRandomPostalCodes("./../../datasets/zipCodes_small.csv", 20)
	require.NoError(suite.T(), err)

	for _, code := range postalCodes {
		suite.Run(code, func() {
			req := httptest.NewRequest("GET", "/postalcode?postalcode="+url.QueryEscape(code), nil)
			resp, _ := suite.app.Test(req, -1)

			if resp.StatusCode != http.StatusOK {
				suite.T().Logf("Postal code %s not found, skipping...", code)
				return
			}

			assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

			var cities []*city.City
			err := json.NewDecoder(resp.Body).Decode(&cities)
			assert.NoError(suite.T(), err)
			assert.NotEmpty(suite.T(), cities)
		})
	}
}

func (suite *ServerTestSuite) TestBadRequest() {
	suite.T().Skip("Skipping test due to OOM issues")
	req := httptest.NewRequest("GET", "/nearest?lat=invalid&lon=-74.0060", nil)
	resp, _ := suite.app.Test(req, -1)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(suite.T(), err)
	bodyString := string(bodyBytes)

	assert.Equal(suite.T(), "Invalid latitude", bodyString)
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
