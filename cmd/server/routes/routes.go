// cmd/server/routes.go
package routes

import (
	"fmt"
	"github.com/SamyRai/cityFinder/lib/finder"
	"github.com/gofiber/fiber/v2"
	"log"
	"strconv"
	"strings"
)

func SetupRoutes(app *fiber.App, mainFinder *finder.Finder) {
	app.Get("/nearest", func(c *fiber.Ctx) error {
		lat, err := strconv.ParseFloat(c.Query("lat"), 64)
		if err != nil {
			log.Printf("Error parsing lat: %v", err)
			return c.Status(fiber.StatusBadRequest).SendString("Invalid latitude")
		}
		lon, err := strconv.ParseFloat(c.Query("lon"), 64)
		if err != nil {
			log.Printf("Error parsing lon: %v", err)
			return c.Status(fiber.StatusBadRequest).SendString("Invalid longitude")
		}

		if lat < -90 || lat > 90 {
			return c.Status(fiber.StatusBadRequest).SendString("Latitude must be between -90 and 90")
		}

		if lon < -180 || lon > 180 {
			return c.Status(fiber.StatusBadRequest).SendString("Longitude must be between -180 and 180")
		}

		city, _, err := mainFinder.FindNearestCity(lat, lon)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error finding city: %v", err))
		}
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
		countryCode := strings.ToUpper(c.Query("country-code"))
		if countryCode == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Country code is required")
		}

		city := mainFinder.FindCityByName(name, countryCode)
		if city == nil {
			return c.Status(fiber.StatusNotFound).SendString("City not found")
		}

		return c.JSON(city)
	})

	app.Get("/postalCode", func(c *fiber.Ctx) error {
		postalCode := c.Query("code")
		countryCode := strings.ToUpper(c.Query("country-code"))
		if postalCode == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Postal code is required")
		}
		if countryCode == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Country code is required")
		}
		city := mainFinder.FindCityByPostalCode(postalCode, countryCode)
		if city == nil {
			return c.Status(fiber.StatusNotFound).SendString("City not found")
		}

		return c.JSON(city)
	})
}
