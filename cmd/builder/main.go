package main

import (
	"github.com/SamyRai/cityFinder/dataloader"
	"github.com/SamyRai/cityFinder/finder"
	"log"
)

import "flag"

func main() {
	var (
		csvLocation        = flag.String("cities", "datasets/cities.csv", "Path to the cities CSV file")
		postalCodeLocation = flag.String("postal", "datasets/zipcodes.csv", "Path to the postal codes CSV file")
		finderType         = flag.String("finder", "all", "Finder type to build (kdtree, rtree, geohash, all)")
	)
	flag.Parse()

	log.Printf("Loading the CSV data from %s", *csvLocation)
	cities, err := dataloader.LoadGeoNamesCSV(*csvLocation)
	if err != nil {
		log.Fatalf("Failed to load GeoNames data from CSV: %v", err)
	}
	log.Printf("Finished loading CSV data")

	log.Printf("Loading the Postal Code data from %s", *postalCodeLocation)
	postalCodes, err := dataloader.LoadPostalCodes(*postalCodeLocation)
	if err != nil {
		log.Fatalf("Failed to load postal codes: %v", err)
	}
	log.Printf("Finished loading Postal Code data")

	if *finderType == "all" || *finderType == "kdtree" {
		log.Println("Building k-d tree index")
		finder.BuildKDTree(cities, postalCodes)
	}
	if *finderType == "all" || *finderType == "rtree" {
		log.Println("Building R-tree index")
		finder.BuildRTree(cities, postalCodes)
	}
	if *finderType == "all" || *finderType == "geohash" {
		log.Println("Building geohash index")
		finder.BuildGeoHashIndex(cities, 12, postalCodes)
	}
}
