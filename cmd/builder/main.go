package main

import (
	"github.com/SamyRai/cityFinder/dataloader"
	"github.com/SamyRai/cityFinder/finder"
	"log"
)

func main() {
	csvLocation := "datasets/allCountries.csv"
	postalCodeLocation := "datasets/zipCodes.csv"
	indexLocation := "datasets/s2index.gob"

	log.Printf("Loading the CSV data from %s", csvLocation)
	cities, err := dataloader.LoadGeoNamesCSV(csvLocation)
	if err != nil {
		log.Fatalf("Failed to load GeoNames data from CSV: %v", err)
	}
	log.Printf("Finished loading CSV data")

	log.Printf("Loading the Postal Code data from %s", postalCodeLocation)
	postalCodes, err := dataloader.LoadPostalCodes(postalCodeLocation)
	if err != nil {
		log.Fatalf("Failed to load Postal Code data: %v", err)
	}
	log.Printf("Finished loading Postal Code data")

	log.Printf("Building S2 index")
	s2Finder := finder.BuildS2Index(cities, postalCodes)

	log.Printf("Serializing S2 index to %s", indexLocation)
	err = s2Finder.SerializeIndex(indexLocation)
	if err != nil {
		log.Fatalf("Failed to serialize S2 index: %v", err)
	}
	log.Printf("Finished serializing S2 index")
}
