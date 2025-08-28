// cmd/server/routes.go
package main

import (
	"fmt"
	"github.com/SamyRai/cityFinder/finder"
	"github.com/gofiber/fiber/v2"
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

		if lat < -90 || lat > 90 {
			return c.Status(fiber.StatusBadRequest).SendString("Latitude must be between -90 and 90")
		}

		if lon < -180 || lon > 180 {
			return c.Status(fiber.StatusBadRequest).SendString("Longitude must be between -180 and 180")
		}

		city := f.FindNearestCity(lat, lon)
		if city == nil {
			return c.Status(fiber.StatusNotFound).SendString(fmt.Sprintf("City not found for lat: %f, lon: %f", lat, lon))
		}
		return c.JSON(city)
	})

	app.Get("/coordinates", func(c *fiber.Ctx) error {
		name := c.Query("name")
		if name == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Name is required")
		}

		cities := f.FindCoordinatesByName(name)
		if cities == nil {
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
		if cities == nil {
			return c.Status(fiber.StatusNotFound).SendString("City not found")
		}

		return c.JSON(cities)
	})
}
