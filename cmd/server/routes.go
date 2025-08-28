// cmd/server/routes.go
package main

import (
	"errors"
	"fmt"
	"github.com/SamyRai/cityFinder/finder"
	"github.com/gofiber/fiber/v2"
	"log"
	"strconv"
)

func setupRoutes(app *fiber.App, f finder.Finder) {
	app.Get("/nearest", func(c *fiber.Ctx) error {
		lat, err := strconv.ParseFloat(c.Query("lat"), 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid latitude")
		}
		lon, err := strconv.ParseFloat(c.Query("lon"), 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid longitude")
		}

		city, _, err := f.FindNearestCity(lat, lon)
		if err != nil {
			if errors.Is(err, finder.ErrNoResults) || errors.Is(err, finder.ErrIndexOutOfRange) {
				return c.Status(fiber.StatusNotFound).SendString(fmt.Sprintf("City not found for lat: %f, lon: %f", lat, lon))
			}
			if errors.Is(err, finder.ErrOutOfRange) {
				return c.Status(fiber.StatusBadRequest).SendString(err.Error())
			}
			log.Printf("Error finding nearest city: %v", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Error finding nearest city")
		}

		return c.JSON(city)
	})

	app.Get("/coordinates", func(c *fiber.Ctx) error {
		name := c.Query("name")
		if name == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Name is required")
		}

		cities := f.FindCoordinatesByName(name)
		if len(cities) == 0 {
			return c.Status(fiber.StatusNotFound).SendString("City not found")
		}

		return c.JSON(cities)
	})

	app.Get("/postalcode", func(c *fiber.Ctx) error {
		postalCode := c.Query("postalcode")
		if postalCode == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Postal code is required")
		}
		cities := f.FindCityByPostalCode(postalCode)
		if len(cities) == 0 {
			return c.Status(fiber.StatusNotFound).SendString("City not found")
		}

		return c.JSON(cities)
	})
}
