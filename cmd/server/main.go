package main

import (
	"encoding/json"
	"fmt"
	"github.com/SamyRai/cityFinder/cmd/server/routes"
	"github.com/SamyRai/cityFinder/lib/config"
	"github.com/SamyRai/cityFinder/lib/initializer"
	"github.com/fatih/color"
	"github.com/gofiber/fiber/v2"
	"log"
	"os"
	"time"
)

func main() {
	configPath, exists := os.LookupEnv("CONFIG_PATH")
	if !exists {
		configPath = "config.json"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	mainFinder, err := initializer.Initialize(cfg)
	if err != nil {
		log.Fatalf("Initialization failed: %v", err)
	}

	app := fiber.New(fiber.Config{
		ETag:              true,
		EnablePrintRoutes: true,
	})
	app.Use(Logger())
	routes.SetupRoutes(app, mainFinder)

	log.Fatal(app.Listen(":3000"))
}

func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		stop := time.Now()

		// Color functions
		timeColor := color.New(color.FgCyan).SprintFunc()
		methodColor := color.New(color.FgGreen).SprintFunc()
		pathColor := color.New(color.FgYellow).SprintFunc()
		statusColor := color.New(color.FgRed).SprintFunc()
		latencyColor := color.New(color.FgBlue).SprintFunc()
		paramsColor := color.New(color.FgMagenta).SprintFunc()
		queryColor := color.New(color.FgWhite).SprintFunc()
		bodyColor := color.New(color.FgHiWhite).SprintFunc()

		// Get multipart form data
		form, _ := c.MultipartForm()

		// Convert query parameters to a string
		queryParams, err := json.Marshal(c.Queries())
		if err != nil {
			log.Printf("Error marshalling query params: %v", err)
		}

		// Get the request body
		body := fmt.Sprintf(`"%s"`, c.Body())

		formData := c.Locals("formData")
		formDataString, err := json.Marshal(formData)
		if err != nil {
			log.Printf("Error marshalling form data: %v", err)
		}

		// Log output
		log.Printf(
			"{\n\"time\": \"%s\",\n\"method\": \"%s\",\n\"path\": \"%s\",\n\"status\": %s,\n\"latency\": \"%s\",\n\"params\": %s,\n\"query\": %s,\n\"body\": %s,\n\"formData\": %s\n}",
			timeColor(start.Format(time.RFC3339)),
			methodColor(c.Method()),
			pathColor(c.Path()),
			statusColor(c.Response().StatusCode()),
			latencyColor(stop.Sub(start).String()),
			paramsColor(form),
			queryColor(string(queryParams)),
			bodyColor(body),
			queryColor(string(formDataString)),
		)

		return err
	}
}
