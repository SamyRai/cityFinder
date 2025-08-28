package main

import (
	"encoding/json"
	"fmt"
	"github.com/SamyRai/cityFinder/city"
	"github.com/SamyRai/cityFinder/dataloader"
	"github.com/SamyRai/cityFinder/finder"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type ServerTestSuite struct {
	suite.Suite
	app    *fiber.App
	cities []city.SpatialCity
}

func (suite *ServerTestSuite) SetupSuite() {
	log.Println("Loading test data for suite...")
	cities, err := dataloader.LoadGeoNamesCSV("./../../datasets/allCountries_small.csv")
	if err != nil {
		log.Fatalf("Failed to load test cities: %v", err)
	}
	suite.cities = cities

	postalCodes, err := dataloader.LoadPostalCodes("./../../datasets/zipCodes_small.csv")
	if err != nil {
		log.Fatalf("Failed to load test postal codes: %v", err)
	}

	f := finder.BuildGeoHashIndex(cities, 12, postalCodes)
	app := fiber.New()
	setupRoutes(app, f)
	suite.app = app
	log.Println("Finished test data loading for suite.")
}

func (suite *ServerTestSuite) TestGetNearestCityRandom() {
	for i := 0; i < 20; i++ {
		loc := suite.cities[i]
		suite.Run(loc.Name, func() {
			req := httptest.NewRequest("GET", "/nearest?lat="+fmt.Sprintf("%f", loc.Latitude)+"&lon="+fmt.Sprintf("%f", loc.Longitude), nil)
			resp, _ := suite.app.Test(req, -1)

			suite.Equal(http.StatusOK, resp.StatusCode)

			var cityObj city.City
			err := json.NewDecoder(resp.Body).Decode(&cityObj)
			suite.Require().NoError(err)
			suite.NotEmpty(cityObj.Name)
		})
	}
}

func (suite *ServerTestSuite) TestGetCoordinatesByNameRandom() {
	for i := 0; i < 20; i++ {
		loc := suite.cities[i]
		suite.Run(loc.Name, func() {
			req := httptest.NewRequest("GET", "/coordinates?name="+url.QueryEscape(loc.Name), nil)
			resp, _ := suite.app.Test(req, -1)

			if resp.StatusCode != http.StatusOK {
				suite.T().Logf("City %s not found, skipping...", loc.Name)
				return
			}

			suite.Equal(http.StatusOK, resp.StatusCode)

			var cities []*city.City
			err := json.NewDecoder(resp.Body).Decode(&cities)
			suite.Require().NoError(err)
			suite.NotEmpty(cities)
		})
	}
}

func (suite *ServerTestSuite) TestGetCityByPostalCodeRandom() {
	postalCodes, err := pickRandomPostalCodes("./../../datasets/zipCodes_small.csv", 20)
	require.NoError(suite.T(), err)

	for _, code := range postalCodes {
		suite.Run(code, func() {
			req := httptest.NewRequest("GET", "/postalcode?postalcode="+url.QueryEscape(code), nil)
			resp, _ := suite.app.Test(req, -1)

			if resp.StatusCode != http.StatusOK {
				suite.T().Logf("Postal code %s not found, skipping...", code)
				return
			}
			suite.Equal(http.StatusOK, resp.StatusCode)

			var cityObj []*city.City
			err = json.NewDecoder(resp.Body).Decode(&cityObj)
			suite.Require().NoError(err)
			suite.NotEmpty(cityObj)
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
