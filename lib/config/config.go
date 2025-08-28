package config

import (
	"encoding/json"
	"fmt"
	"github.com/SamyRai/cityFinder/util"
	"os"
	"path/filepath"
)

type Config struct {
	DatasetsFolder      string `json:"datasets_folder"`
	AllCitiesURL        string `json:"all_cities_url"`
	PostalCodesURL      string `json:"postal_codes_url"`
	AllCitiesFile       string `json:"all_cities_file"`
	PostalCodesFile     string `json:"postal_codes_file"`
	AllCitiesZip        string `json:"all_cities_zip"`
	PostalCodesZip      string `json:"postal_codes_zip"`
	NameIndexFile       string `json:"name_index_file"`
	PostalCodeIndexFile string `json:"postal_code_index_file"`
	S2                  S2     `json:"s2"`
}

type S2 struct {
	MinLevel  int    `json:"min_level"`
	MaxLevel  int    `json:"max_level"`
	MaxCells  int    `json:"max_cells"`
	IndexFile string `json:"index_file"`
}

func LoadConfig(configPath string) (*Config, error) {
	cfg := &Config{}

	rootDir, err := util.FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %v", err)
	}

	if configPath == "" {
		configPath = os.Getenv("CONFIG_FILE")
	}

	// If a configPath is provided, load config from that path
	if configPath != "" {
		file, err := os.Open(filepath.Join(rootDir, configPath))
		if err != nil {
			return nil, fmt.Errorf("failed to open config file: %v", err)
		}
		defer file.Close()

		decoder := json.NewDecoder(file)
		err = decoder.Decode(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to decode config file: %v", err)
		}

		cfg.DatasetsFolder = filepath.Join(rootDir, cfg.DatasetsFolder)

		return cfg, nil
	}

	return nil, fmt.Errorf("no config file provided")
}

func getEnv(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
