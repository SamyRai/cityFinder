package initializer

import (
	"archive/zip"
	"fmt"
	"github.com/SamyRai/cityFinder/lib/config"
	"github.com/SamyRai/cityFinder/lib/dataLoader"
	"github.com/SamyRai/cityFinder/lib/finder"
	"github.com/SamyRai/cityFinder/lib/finder/coordinates"
	"github.com/SamyRai/cityFinder/lib/finder/name"
	"github.com/SamyRai/cityFinder/lib/finder/postalCode"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Initialize ensures datasets are downloaded and extracted, and the indexes are built
func Initialize(cfg *config.Config) (*finder.Finder, error) {
	if err := ensureDatasets(cfg); err != nil {
		return nil, err
	}
	return ensureFinders(cfg)
}

// ensureDatasets ensures that the datasets are downloaded and extracted
func ensureDatasets(cfg *config.Config) error {
	log.Printf("Ensuring datasets are downloaded and extracted in %s", cfg.DatasetsFolder)
	if _, err := os.Stat(cfg.DatasetsFolder); os.IsNotExist(err) {
		err := os.Mkdir(cfg.DatasetsFolder, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create datasets folder: %v", err)
		}
	}

	if err := downloadAndExtractDataset(cfg.AllCitiesURL, cfg.AllCitiesZip, cfg.AllCitiesFile, cfg); err != nil {
		return err
	}
	if err := downloadAndExtractDataset(cfg.PostalCodesURL, cfg.PostalCodesZip, cfg.PostalCodesFile, cfg); err != nil {
		return err
	}
	return nil
}

// downloadAndExtractDataset downloads and extracts the dataset if not already present
func downloadAndExtractDataset(url, zipName, fileName string, cfg *config.Config) error {
	zipPath := filepath.Join(cfg.DatasetsFolder, zipName)
	filePath := filepath.Join(cfg.DatasetsFolder, fileName)

	// Check if the final extracted file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Check if the zip file exists before downloading
		if _, err := os.Stat(zipPath); os.IsNotExist(err) {
			log.Printf("Downloading %s...", url)
			err := downloadFile(zipPath, url)
			if err != nil {
				return fmt.Errorf("failed to download %s: %v", url, err)
			}
		} else {
			log.Printf("Zip file %s already exists, skipping download.", zipPath)
		}

		log.Printf("Extracting %s...", zipName)
		err = unzipAndRename(zipPath, cfg.DatasetsFolder, fileName)
		if err != nil {
			return fmt.Errorf("failed to extract %s: %v", zipName, err)
		}
	} else {
		log.Printf("Dataset file %s already exists, skipping extraction.", filePath)
	}
	return nil
}

// downloadFile downloads a file from a given URL
func downloadFile(filepath string, url string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// unzipAndRename unzips a file and renames it to the specified new file name
func unzipAndRename(src string, dest string, newFileName string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", fpath)
		}
		if f.FileInfo().IsDir() {
			err := os.MkdirAll(fpath, os.ModePerm)
			if err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}
		outFile, err := os.OpenFile(filepath.Join(dest, newFileName), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		_, err = io.Copy(outFile, rc)
		err = outFile.Close()
		if err != nil {
			return fmt.Errorf("failed to close file: %v", err)
		}
		err = rc.Close()
		if err != nil {
			return fmt.Errorf("failed to close file: %v", err)
		}
	}
	return nil
}

// ensureFinders ensures that the indexes are built and serialized
func ensureFinders(cfg *config.Config) (*finder.Finder, error) {
	s2IndexPath := filepath.Join(cfg.DatasetsFolder, cfg.S2.IndexFile)
	nameIndexPath := filepath.Join(cfg.DatasetsFolder, cfg.NameIndexFile)
	postalCodeIndexPath := filepath.Join(cfg.DatasetsFolder, cfg.PostalCodeIndexFile)

	var s2Finder *coordinates.S2Finder
	var nameFinder *name.Finder
	var postalCodeFinder *postalCode.Finder
	var err error

	// Load city and postal code data
	cities, err := dataLoader.LoadGeoNamesCSV(filepath.Join(cfg.DatasetsFolder, cfg.AllCitiesFile))
	if err != nil {
		return nil, fmt.Errorf("failed to load GeoNames data from CSV: %v", err)
	}

	postalCodes, err := dataLoader.LoadPostalCodes(filepath.Join(cfg.DatasetsFolder, cfg.PostalCodesFile))
	if err != nil {
		return nil, fmt.Errorf("failed to load Postal Code data: %v", err)
	}

	// Use a wait group to parallelize the index building
	var wg sync.WaitGroup
	wg.Add(3) // We have three tasks to perform concurrently

	go func() {
		defer wg.Done()
		// Ensure S2 index
		log.Printf("Ensuring S2 index is built and serialized in %s", s2IndexPath)
		if _, err := os.Stat(s2IndexPath); os.IsNotExist(err) {
			log.Printf("S2 index not found in %s\nBuilding it...", s2IndexPath)
			s2Finder, err = coordinates.BuildIndex(cities, &cfg.S2)
			if err != nil {
				log.Fatalf("failed to build S2 index: %v", err)
			}
			err = s2Finder.SerializeIndex(s2IndexPath)
			if err != nil {
				log.Fatalf("failed to serialize S2 index: %v", err)
			}
		} else {
			s2Finder, err = coordinates.DeserializeIndex(s2IndexPath)
			if err != nil {
				log.Fatalf("failed to deserialize S2 index: %v", err)
			}
		}
	}()

	go func() {
		defer wg.Done()
		// Ensure name index
		log.Printf("Ensuring name index is built and serialized in %s", nameIndexPath)
		if _, err := os.Stat(nameIndexPath); os.IsNotExist(err) {
			log.Printf("Name index not found in %s\nBuilding it...", nameIndexPath)
			nameFinder = name.BuildIndex(cities)
			err = nameFinder.SerializeIndex(nameIndexPath)
			if err != nil {
				log.Fatalf("failed to serialize name index: %v", err)
			}
		} else {
			nameFinder, err = name.DeserializeIndex(nameIndexPath)
			if err != nil {
				log.Fatalf("failed to deserialize name index: %v", err)
			}
		}
	}()

	go func() {
		defer wg.Done()
		// Ensure postal code index
		log.Printf("Ensuring postal code index is built and serialized in %s", postalCodeIndexPath)
		if _, err := os.Stat(postalCodeIndexPath); os.IsNotExist(err) {
			log.Printf("Postal code index not found in %s\nBuilding it...", postalCodeIndexPath)
			postalCodeFinder = postalCode.BuildIndex(postalCodes)
			err = postalCodeFinder.SerializeIndex(postalCodeIndexPath)
			if err != nil {
				log.Fatalf("failed to serialize postal code index: %v", err)
			}
		} else {
			postalCodeFinder, err = postalCode.DeserializeIndex(postalCodeIndexPath)
			if err != nil {
				log.Fatalf("failed to deserialize postal code index: %v", err)
			}
		}
	}()

	// Wait for all goroutines to finish
	wg.Wait()

	return &finder.Finder{
		S2Finder:         s2Finder,
		NameFinder:       nameFinder,
		PostalCodeFinder: postalCodeFinder,
	}, nil
}
