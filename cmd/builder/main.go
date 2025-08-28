package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/SamyRai/cityFinder/city"
	"github.com/SamyRai/cityFinder/dataloader"
	"github.com/SamyRai/cityFinder/finder"
	"github.com/dhconnelly/rtreego"
	"github.com/golang/geo/s2"
	"github.com/mmcloughlin/geohash"
)

func main() {
	finderType := flag.String("finder", "s2", "type of finder to build (s2, kdtree, rtree, geohash)")
	csvLocation := flag.String("csv", "datasets/allCountries.csv", "path to allCountries.csv")
	citiesLocation := flag.String("cities", "", "output path for cities gob file")
	pointsLocation := flag.String("points", "", "output path for points gob file (s2 only)")
	metaLocation := flag.String("meta", "", "output path for meta gob file")
	flag.Parse()

	if *citiesLocation == "" {
		*citiesLocation = fmt.Sprintf("datasets/cities_%s.gob", *finderType)
	}
	if *metaLocation == "" {
		*metaLocation = fmt.Sprintf("datasets/meta_%s.gob", *finderType)
	}

	switch *finderType {
	case "s2":
		if *pointsLocation == "" {
			*pointsLocation = fmt.Sprintf("datasets/points_%s.gob", *finderType)
		}
		buildS2(*csvLocation, *citiesLocation, *pointsLocation, *metaLocation)
	case "kdtree":
		buildKDTree(*csvLocation, *citiesLocation, *metaLocation)
	case "rtree":
		buildRTree(*csvLocation, *citiesLocation, *metaLocation)
	case "geohash":
		buildGeoHash(*csvLocation, *citiesLocation, *metaLocation)
	default:
		log.Fatalf("Unknown finder type: %s", *finderType)
	}
}

func buildS2(csvLocation, citiesLocation, pointsLocation, metaLocation string) {
	cityChan, errChan := streamCities(csvLocation)

	citiesFile, pointsFile := createFiles(citiesLocation, pointsLocation)
	defer citiesFile.Close()
	defer pointsFile.Close()
	pointsEncoder := gob.NewEncoder(pointsFile)

	var cityOffsets, cityLengths, pointOffsets []int64

	log.Println("Processing cities for S2...")
	for c := range cityChan {
		point := s2.PointFromLatLng(s2.LatLngFromDegrees(c.Latitude, c.Longitude))

		off, len := writeCity(citiesFile, c)
		cityOffsets = append(cityOffsets, off)
		cityLengths = append(cityLengths, len)

		pointOffset, _ := pointsFile.Seek(0, io.SeekCurrent)
		pointOffsets = append(pointOffsets, pointOffset)
		if err := pointsEncoder.Encode(point); err != nil {
			log.Fatalf("Failed to encode point for city %s: %v", c.Name, err)
		}
	}
	handleStreamError(errChan)

	meta := finder.S2Meta{
		CityOffsets:  cityOffsets,
		CityLengths:  cityLengths,
		PointOffsets: pointOffsets,
	}
	serializeMeta(metaLocation, meta)
}

func buildKDTree(csvLocation, citiesLocation, metaLocation string) {
	cityChan, errChan := streamCities(csvLocation)
	citiesFile, _ := createFiles(citiesLocation, "")
	defer citiesFile.Close()

	var cityOffsets, cityLengths []int64
	var kdTreePoints []finder.KDTreePoint
	var cityCounter int

	log.Println("Processing cities for k-d tree...")
	for c := range cityChan {
		off, len := writeCity(citiesFile, c)
		cityOffsets = append(cityOffsets, off)
		cityLengths = append(cityLengths, len)
		kdTreePoints = append(kdTreePoints, finder.KDTreePoint{
			Coordinates: []float64{c.Longitude, c.Latitude},
			CityID:      cityCounter,
		})
		cityCounter++
	}
	handleStreamError(errChan)

	meta := finder.KDTreeMeta{
		Points:      kdTreePoints,
		CityOffsets: cityOffsets,
		CityLengths: cityLengths,
	}
	serializeMeta(metaLocation, meta)
}

func buildRTree(csvLocation, citiesLocation, metaLocation string) {
	cityChan, errChan := streamCities(csvLocation)
	citiesFile, _ := createFiles(citiesLocation, "")
	defer citiesFile.Close()

	var cityOffsets, cityLengths []int64
	var rTreeSpatials []finder.RTreeSpatial
	var cityCounter int

	log.Println("Processing cities for R-tree...")
	for c := range cityChan {
		off, len := writeCity(citiesFile, c)
		cityOffsets = append(cityOffsets, off)
		cityLengths = append(cityLengths, len)
		rect, _ := rtreego.NewRect(rtreego.Point{c.Longitude, c.Latitude}, []float64{0.00001, 0.00001})
		rTreeSpatials = append(rTreeSpatials, finder.RTreeSpatial{
			Rect:   finder.FromRTreeRect(rect),
			CityID: cityCounter,
		})
		cityCounter++
	}
	handleStreamError(errChan)

	meta := finder.RTreeMeta{
		Spatials:    rTreeSpatials,
		CityOffsets: cityOffsets,
		CityLengths: cityLengths,
	}
	serializeMeta(metaLocation, meta)
}

func buildGeoHash(csvLocation, citiesLocation, metaLocation string) {
	cityChan, errChan := streamCities(csvLocation)
	citiesFile, _ := createFiles(citiesLocation, "")
	defer citiesFile.Close()

	var cityOffsets, cityLengths []int64
	geoHashData := make(map[string][]int)
	var cityCounter int

	log.Println("Processing cities for Geohash...")
	for c := range cityChan {
		off, len := writeCity(citiesFile, c)
		cityOffsets = append(cityOffsets, off)
		cityLengths = append(cityLengths, len)
		hash := geohash.EncodeWithPrecision(c.Latitude, c.Longitude, 12)
		geoHashData[hash] = append(geoHashData[hash], cityCounter)
		cityCounter++
	}
	handleStreamError(errChan)

	meta := finder.GeoHashMeta{
		Data:        geoHashData,
		CityOffsets: cityOffsets,
		CityLengths: cityLengths,
	}
	serializeMeta(metaLocation, meta)
}

// Helper functions
func streamCities(csvLocation string) (<-chan city.SpatialCity, <-chan error) {
	cityChan := make(chan city.SpatialCity)
	errChan := make(chan error, 1)
	go dataloader.StreamGeoNamesCSV(csvLocation, cityChan, errChan)
	return cityChan, errChan
}

func createFiles(citiesPath, pointsPath string) (*os.File, *os.File) {
	citiesFile, err := os.Create(citiesPath)
	if err != nil {
		log.Fatalf("Failed to create cities file: %v", err)
	}
	var pointsFile *os.File
	if pointsPath != "" {
		pointsFile, err = os.Create(pointsPath)
		if err != nil {
			log.Fatalf("Failed to create points file: %v", err)
		}
	}
	return citiesFile, pointsFile
}

func writeCity(file *os.File, c city.SpatialCity) (int64, int64) {
	var cityBuf bytes.Buffer
	if err := gob.NewEncoder(&cityBuf).Encode(finder.FromSpatialCity(c)); err != nil {
		log.Fatalf("Failed to encode city %s: %v", c.Name, err)
	}
	cityBytes := cityBuf.Bytes()

	var lenBuf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(lenBuf[:], uint64(len(cityBytes)))

	offset, _ := file.Seek(0, io.SeekCurrent)
	if _, err := file.Write(lenBuf[:n]); err != nil {
		log.Fatalf("Failed to write length prefix: %v", err)
	}
	if _, err := file.Write(cityBytes); err != nil {
		log.Fatalf("Failed to write city data: %v", err)
	}
	return offset, int64(n + len(cityBytes))
}

func handleStreamError(errChan <-chan error) {
	if err := <-errChan; err != nil {
		log.Fatalf("Error while streaming cities: %v", err)
	}
}

func serializeMeta(path string, meta interface{}) {
	log.Printf("Serializing meta to %s", path)
	metaFile, err := os.Create(path)
	if err != nil {
		log.Fatalf("Failed to create meta file: %v", err)
	}
	defer metaFile.Close()
	if err := gob.NewEncoder(metaFile).Encode(meta); err != nil {
		log.Fatalf("Failed to encode meta file: %v", err)
	}
	log.Println("Finished serializing meta.")
}
