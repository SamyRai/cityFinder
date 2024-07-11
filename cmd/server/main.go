// cmd/server/main.go
package main

import (
	"github.com/SamyRai/cityFinder/finder"
	"github.com/gofiber/fiber/v2"
	"log"
)

func main() {
	indexLocation := "datasets/s2index.gob"

	log.Printf("Loading S2 index from %s", indexLocation)
	s2Finder, err := finder.DeserializeIndex(indexLocation)
	if err != nil {
		log.Fatalf("Failed to load S2 index: %v", err)
	}
	log.Printf("Finished loading S2 index")

	app := fiber.New()
	setupRoutes(app, s2Finder)

	log.Fatal(app.Listen(":3000"))
}
