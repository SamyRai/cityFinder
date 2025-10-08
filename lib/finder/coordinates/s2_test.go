package coordinates

import (
	"os"
	"testing"

	"github.com/SamyRai/cityFinder/lib/city"
	"github.com/SamyRai/cityFinder/lib/config"
	"github.com/stretchr/testify/assert"
)

var testCities = []city.SpatialCity{
	{City: city.City{Name: "San Francisco", Latitude: 37.7749, Longitude: -122.4194}},
	{City: city.City{Name: "New York", Latitude: 40.7128, Longitude: -74.0060}},
	{City: city.City{Name: "London", Latitude: 51.5074, Longitude: -0.1278}},
}

func TestBuildIndex(t *testing.T) {
	cfg := &config.S2{}
	finder, err := BuildIndex(testCities, cfg)

	assert.NoError(t, err)
	assert.NotNil(t, finder)
	assert.NotNil(t, finder.Index)
	assert.Len(t, finder.Cities, 3)
	assert.Equal(t, "San Francisco", finder.Cities[0].Name)
}

func TestNearestPlace(t *testing.T) {
	cfg := &config.S2{}
	finder, _ := BuildIndex(testCities, cfg)

	// Test case 1: Find city closest to SF
	sfLat, sfLon := 37.7750, -122.4190
	nearest, dist, err := finder.NearestPlace(sfLat, sfLon)
	assert.NoError(t, err)
	assert.NotNil(t, nearest)
	assert.Equal(t, "San Francisco", nearest.Name)
	assert.InDelta(t, 0.04, dist, 0.1) // Looser delta for distance

	// Test case 2: Find city closest to NYC
	nycLat, nycLon := 40.7128, -74.0060
	nearest, dist, err = finder.NearestPlace(nycLat, nycLon)
	assert.NoError(t, err)
	assert.NotNil(t, nearest)
	assert.Equal(t, "New York", nearest.Name)
	assert.InDelta(t, 0.0, dist, 0.1)

	// Test case 3: A point in the middle of the Atlantic
	midAtlanticLat, midAtlanticLon := 30.0, -40.0
	nearest, _, err = finder.NearestPlace(midAtlanticLat, midAtlanticLon)
	assert.NoError(t, err)
	assert.NotNil(t, nearest)
	assert.Equal(t, "New York", nearest.Name)
}

func TestSerialization(t *testing.T) {
	// Create a temporary file for the index
	tmpfile, err := os.CreateTemp("", "s2index_test.*.gob")
	assert.NoError(t, err)
	defer func() {
		_ = os.Remove(tmpfile.Name())
	}()

	// Build the initial finder and serialize it
	cfg := &config.S2{}
	finder, _ := BuildIndex(testCities, cfg)
	err = finder.SerializeIndex(tmpfile.Name())
	assert.NoError(t, err)

	// Deserialize the finder
	deserializedFinder, err := DeserializeIndex(tmpfile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, deserializedFinder)
	assert.Len(t, deserializedFinder.Cities, 3)

	// Test that the deserialized finder works correctly
	sfLat, sfLon := 37.7750, -122.4190
	nearest, _, err := deserializedFinder.NearestPlace(sfLat, sfLon)
	assert.NoError(t, err)
	assert.NotNil(t, nearest)
	assert.Equal(t, "San Francisco", nearest.Name)
}

func TestEmptyCities(t *testing.T) {
	cfg := &config.S2{}
	finder, err := BuildIndex([]city.SpatialCity{}, cfg)
	assert.NoError(t, err)
	assert.NotNil(t, finder)

	_, _, err = finder.NearestPlace(0, 0)
	assert.Error(t, err)
	assert.Equal(t, "no city found", err.Error())
}

func TestSingleCity(t *testing.T) {
	cfg := &config.S2{}
	singleCityList := []city.SpatialCity{
		{City: city.City{Name: "Honolulu", Latitude: 21.3069, Longitude: -157.8583}},
	}
	finder, err := BuildIndex(singleCityList, cfg)
	assert.NoError(t, err)

	nearest, _, err := finder.NearestPlace(21.3, -157.8)
	assert.NoError(t, err)
	assert.NotNil(t, nearest)
	assert.Equal(t, "Honolulu", nearest.Name)
}