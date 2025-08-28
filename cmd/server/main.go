package main

import (
	"flag"
	"fmt"
	"github.com/SamyRai/cityFinder/finder"
	"github.com/gofiber/fiber/v2"
	"log"
)

func main() {
	finderType := flag.String("finder", "s2", "type of finder to use (s2, kdtree, rtree, geohash)")
	flag.Parse()

	var f finder.Finder
	var err error

	citiesPath := fmt.Sprintf("datasets/cities_%s.gob", *finderType)
	metaPath := fmt.Sprintf("datasets/meta_%s.gob", *finderType)
	pointsPath := fmt.Sprintf("datasets/points_%s.gob", *finderType)

	log.Printf("Loading %s index...", *finderType)
	switch *finderType {
	case "s2":
		f, err = finder.DeserializeS2(metaPath, citiesPath, pointsPath)
	case "kdtree":
		f, err = finder.DeserializeKDTree(metaPath, citiesPath)
	case "rtree":
		f, err = finder.DeserializeRTree(metaPath, citiesPath)
	case "geohash":
		f, err = finder.DeserializeGeoHash(metaPath, citiesPath)
	default:
		log.Fatalf("Unknown finder type: %s", *finderType)
	}

	if err != nil {
		log.Fatalf("Failed to load index: %v", err)
	}
	defer f.Close()
	log.Printf("Finished loading %s index", *finderType)

	app := fiber.New()
	setupRoutes(app, f)

	log.Fatal(app.Listen(":3000"))
}
