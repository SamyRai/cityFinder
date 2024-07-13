// cmd/server/routes.go
package routes

import (
	"fmt"
	"github.com/SamyRai/cityFinder/lib/finder"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

func SetupRoutes(app *fiber.App, s2Finder *finder.S2Finder) {
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

		city := s2Finder.FindNearestCity(lat, lon)
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

		city := s2Finder.FindCoordinatesByName(name)
		if city == nil {
			return c.Status(fiber.StatusNotFound).SendString("City not found")
		}

		return c.JSON(city)
	})

	app.Get("/postalcode", func(c *fiber.Ctx) error {
		postalCode := c.Query("postalcode")
		if postalCode == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Postal code is required")
		}
		city := s2Finder.FindCityByPostalCode(postalCode)
		if city == nil {
			return c.Status(fiber.StatusNotFound).SendString("City not found")
		}

		return c.JSON(city)
	})
}
