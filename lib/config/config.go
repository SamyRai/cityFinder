package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	DatasetsFolder  string `json:"datasets_folder"`
	AllCitiesURL    string `json:"all_cities_url"`
	PostalCodesURL  string `json:"postal_codes_url"`
	AllCitiesFile   string `json:"all_cities_file"`
	PostalCodesFile string `json:"postal_codes_file"`
	AllCitiesZip    string `json:"all_cities_zip"`
	PostalCodesZip  string `json:"postal_codes_zip"`
	S2IndexFile     string `json:"s2_index_file"`
}

var cfg *Config

func LoadConfig(configPath string) (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		// Default configuration values
		DatasetsFolder:  getEnv("DATASETS_FOLDER", "datasets"),
		AllCitiesURL:    getEnv("ALL_CITIES_URL", "http://download.geonames.org/export/dump/allCountries.zip"),
		PostalCodesURL:  getEnv("POSTAL_CODES_URL", "http://download.geonames.org/export/zip/allCountries.zip"),
		AllCitiesFile:   getEnv("ALL_CITIES_FILE", "allCountries.txt"),
		PostalCodesFile: getEnv("POSTAL_CODES_FILE", "zipCodes.txt"),
		AllCitiesZip:    getEnv("ALL_CITIES_ZIP", "allCountries.zip"),
		PostalCodesZip:  getEnv("POSTAL_CODES_ZIP", "zipCodes.zip"),
		S2IndexFile:     getEnv("S2_INDEX_FILE", "datasets/s2index.gob"),
	}

	// If a configPath is provided, load config from that path
	if configPath != "" {
		file, err := os.Open(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open config file: %v", err)
		}
		defer file.Close()

		decoder := json.NewDecoder(file)
		err = decoder.Decode(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to decode config file: %v", err)
		}
		return cfg, nil
	}

	// If CONFIG_FILE environment variable is set, override with that file
	configFile := os.Getenv("CONFIG_FILE")
	if configFile != "" {
		file, err := os.Open(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to open config file from CONFIG_FILE env: %v", err)
		}
		defer file.Close()

		decoder := json.NewDecoder(file)
		err = decoder.Decode(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to decode config file from CONFIG_FILE env: %v", err)
		}
	}

	return cfg, nil
}

func getEnv(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
