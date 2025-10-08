package name

import (
	"github.com/SamyRai/cityFinder/lib/city"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestFinder_SerializeDeserialize(t *testing.T) {
	// Create a new finder and add some data
	finder := NewNameFinder()
	finder.AddCity(city.SpatialCity{
		City: city.City{
			Name:    "Test City",
			Country: "TC",
		},
	})

	// Serialize the finder to a temporary file
	tmpfile, err := os.CreateTemp("", "test_name_finder_*.gob")
	assert.NoError(t, err)
	defer func() {
		_ = os.Remove(tmpfile.Name())
	}()

	err = finder.SerializeIndex(tmpfile.Name())
	assert.NoError(t, err)

	// Deserialize the finder from the file
	deserializedFinder, err := DeserializeIndex(tmpfile.Name())
	assert.NoError(t, err)

	// Compare the original and deserialized finders
	assert.Equal(t, finder.InvertedIndex, deserializedFinder.InvertedIndex)
	assert.NotNil(t, deserializedFinder.BKTree)
	assert.Equal(t, finder.BKTree.Root.Term, deserializedFinder.BKTree.Root.Term)
}
