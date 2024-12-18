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

type ServerTestSuite struct {
	suite.Suite
	app      *fiber.App
	s2Finder *finder.S2Finder
}

func (suite *ServerTestSuite) SetupSuite() {
	s2indexPath := "./../../datasets/s2index.gob"
	fmt.Printf("Loading S2 index from %s", s2indexPath)

	s2Finder, err := finder.DeserializeIndex(s2indexPath)
	require.NoError(suite.T(), err)

	suite.s2Finder = s2Finder
	suite.app = setupMockAppTestify(s2Finder)
}

func setupMockAppTestify(s2Finder *finder.S2Finder) *fiber.App {
	app := fiber.New()
	setupRoutes(app, s2Finder)
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
	lines, err := suite.pickRandomLines("./../../datasets/allCountries.csv", 20)
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
	lines, err := suite.pickRandomLines("./../../datasets/allCountries.csv", 20)
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

			var cityObj city.City
			err := json.NewDecoder(resp.Body).Decode(&cityObj)
			assert.NoError(suite.T(), err)
			assert.Equal(suite.T(), loc.Latitude, cityObj.Latitude)
			assert.Equal(suite.T(), loc.Longitude, cityObj.Longitude)
		})
	}
}

func (suite *ServerTestSuite) TestGetCityByPostalCodeRandom() {
	postalCodes, err := suite.pickRandomPostalCodes("./../../datasets/zipCodes.csv", 20)
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

			var cityObj city.City
			err := json.NewDecoder(resp.Body).Decode(&cityObj)
			assert.NoError(suite.T(), err)
			assert.NotEmpty(suite.T(), cityObj.Name)
		})
	}
}

func (suite *ServerTestSuite) TestBadRequest() {
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
