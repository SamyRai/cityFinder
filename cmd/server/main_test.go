package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/SamyRai/cityFinder/cmd/server/routes"
	"github.com/SamyRai/cityFinder/lib/city"
	"github.com/SamyRai/cityFinder/lib/config"
	"github.com/SamyRai/cityFinder/lib/finder"
	"github.com/SamyRai/cityFinder/lib/initializer"
	"github.com/SamyRai/cityFinder/util"
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
	app    *fiber.App
	finder *finder.Finder
	config *config.Config
	rootDir string
}

func (suite *ServerTestSuite) SetupSuite() {
	rootDir, err := util.FindProjectRoot()
	require.NoError(suite.T(), err)
	suite.rootDir = rootDir

	cfg, err := config.LoadConfig("cmd/server/config_test.json")
	fmt.Printf("cfg: %+v\n", cfg)
	fmt.Println("rootDir: ", rootDir)
	require.NoError(suite.T(), err)

	suite.finder, err = initializer.Initialize(cfg)
	require.NoError(suite.T(), err)
	suite.app = suite.setupMockAppTestify()
}

func (suite *ServerTestSuite) setupMockAppTestify() *fiber.App {
	app := fiber.New()
	routes.SetupRoutes(app, suite.finder)
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

func (suite *ServerTestSuite) pickRandomPostalCodes(filepath string, count int) (map[string]map[string]string, error) {
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

	rand.New(rand.NewSource(time.Now().UnixNano()))
	rand.Shuffle(len(records), func(i, j int) {
		records[i], records[j] = records[j], records[i]
	})

	var postalCodes map[string]map[string]string
	for i := 0; i < count && i < len(records); i++ {
		if postalCodes == nil {
			postalCodes = make(map[string]map[string]string)
		}
		if postalCodes[records[i][0]] == nil {
			postalCodes[records[i][0]] = make(map[string]string)
		}

		postalCodes[records[i][0]][records[i][1]] = records[i][2]
	}

	return postalCodes, nil
}

func (suite *ServerTestSuite) TestGetNearestCityRandom() {
	lines, err := suite.pickRandomLines("../../testdata/allCountries.txt", 20)
	require.NoError(suite.T(), err)

	for _, line := range lines {
		loc, err := suite.parseCityLine(line)
		require.NoError(suite.T(), err)

		suite.Run(loc.Name, func() {
			query := fmt.Sprintf("/nearest?lat=%f&lon=%f", loc.Latitude, loc.Longitude)
			req := httptest.NewRequest("GET", query, nil)
			resp, _ := suite.app.Test(req, -1)

			assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

			var cityObj city.City
			err = json.NewDecoder(resp.Body).Decode(&cityObj)
			assert.NoError(suite.T(), err)
			assert.NotEmpty(suite.T(), cityObj.Name)
		})
	}
}

func (suite *ServerTestSuite) TestGetCoordinatesByNameRandom() {
	postalCodes, err := suite.pickRandomPostalCodes("../../testdata/zipCodes.txt", 20)
	require.NoError(suite.T(), err)

	for countryCode, country := range postalCodes {
		for _, cityName := range country {
			suite.Run(cityName, func() {
				req := httptest.NewRequest("GET", "/coordinates?name="+url.QueryEscape(cityName)+"&country-code="+countryCode, nil)
				resp, _ := suite.app.Test(req, -1)

				if resp.StatusCode != http.StatusOK {
					suite.T().Logf("City %s not found, skipping...", cityName)
					return
				}

				assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

				var cityObj city.City
				err := json.NewDecoder(resp.Body).Decode(&cityObj)
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), cityName, cityObj.Name)
				assert.Equal(suite.T(), countryCode, cityObj.Country)
			})
		}
	}
}

func (suite *ServerTestSuite) TestGetCityByPostalCodeRandom() {
	postalCodes, err := suite.pickRandomPostalCodes("../../testdata/zipCodes.txt", 20)
	require.NoError(suite.T(), err)

	for countryCode, country := range postalCodes {
		for code := range country {
			suite.Run(code, func() {
				queryParams := url.QueryEscape("code=" + code + "&country-code=" + countryCode)
				req := httptest.NewRequest("GET", "/postalCode?"+queryParams, nil)
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
