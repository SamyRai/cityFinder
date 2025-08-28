// cmd/server/main.go
package main

import (
	"github.com/SamyRai/cityFinder/finder"
	"github.com/gofiber/fiber/v2"
	"log"
)

import (
	"flag"
	"github.com/SamyRai/cityFinder/dataloader"
)

func main() {
	var (
		citiesPath     = flag.String("cities", "datasets/cities.csv", "Path to the cities CSV file")
		postalCodesPath = flag.String("postal", "datasets/zipcodes.csv", "Path to the postal codes CSV file")
		finderType     = flag.String("finder", "geohash", "Finder type to use (kdtree, rtree, geohash)")
	)
	flag.Parse()

	cities, err := dataloader.LoadGeoNamesCSV(*citiesPath)
	if err != nil {
		log.Fatalf("Failed to load GeoNames data from CSV: %v", err)
	}
	postalCodes, err := dataloader.LoadPostalCodes(*postalCodesPath)
	if err != nil {
		log.Fatalf("Failed to load postal codes: %v", err)
	}

	var f finder.Finder
	switch *finderType {
	case "kdtree":
		f = finder.BuildKDTree(cities, postalCodes)
	case "rtree":
		f = finder.BuildRTree(cities, postalCodes)
	case "geohash":
		f = finder.BuildGeoHashIndex(cities, 12, postalCodes)
	default:
		log.Fatalf("Unknown finder type: %s", *finderType)
	}

	app := fiber.New()
	setupRoutes(app, f)

	log.Fatal(app.Listen(":3000"))
}
