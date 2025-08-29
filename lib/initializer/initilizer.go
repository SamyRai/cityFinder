package initializer

import (
	"archive/zip"
	"fmt"
	"github.com/SamyRai/cityFinder/lib/city"
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
	if zipName == "" {
		return nil
	}
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
	cities, postalCodes, err := loadData(cfg)
	if err != nil {
		return nil, err
	}

	s2Finder, err := ensureS2Index(cfg, cities)
	if err != nil {
		return nil, err
	}

	nameFinder, err := ensureNameIndex(cfg, cities)
	if err != nil {
		return nil, err
	}

	postalCodeFinder, err := ensurePostalCodeIndex(cfg, postalCodes)
	if err != nil {
		return nil, err
	}

	return &finder.Finder{
		S2Finder:         s2Finder,
		NameFinder:       nameFinder,
		PostalCodeFinder: postalCodeFinder,
	}, nil
}

func loadData(cfg *config.Config) ([]city.SpatialCity, map[string]map[string]dataLoader.PostalCodeEntry, error) {
	cities, err := dataLoader.LoadGeoNamesCSV(filepath.Join(cfg.DatasetsFolder, cfg.AllCitiesFile))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load GeoNames data from CSV: %v", err)
	}

	postalCodes, err := dataLoader.LoadPostalCodes(filepath.Join(cfg.DatasetsFolder, cfg.PostalCodesFile))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load Postal Code data: %v", err)
	}

	return cities, postalCodes, nil
}

func ensureS2Index(cfg *config.Config, cities []city.SpatialCity) (*coordinates.S2Finder, error) {
	s2IndexPath := filepath.Join(cfg.DatasetsFolder, cfg.S2.IndexFile)
	var s2Finder *coordinates.S2Finder
	var err error

	log.Printf("Ensuring S2 index is built and serialized in %s", s2IndexPath)
	if _, errStat := os.Stat(s2IndexPath); os.IsNotExist(errStat) {
		log.Printf("S2 index not found in %s\nBuilding it...", s2IndexPath)
		s2Finder, err = coordinates.BuildIndex(cities, &cfg.S2)
		if err != nil {
			return nil, fmt.Errorf("failed to build S2 index: %v", err)
		}
		err = s2Finder.SerializeIndex(s2IndexPath)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize S2 index: %v", err)
		}
	} else {
		s2Finder, err = coordinates.DeserializeIndex(s2IndexPath)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize S2 index: %v", err)
		}
	}
	return s2Finder, nil
}

func ensureNameIndex(cfg *config.Config, cities []city.SpatialCity) (*name.Finder, error) {
	nameIndexPath := filepath.Join(cfg.DatasetsFolder, cfg.NameIndexFile)
	var nameFinder *name.Finder
	var err error

	log.Printf("Ensuring name index is built and serialized in %s", nameIndexPath)
	if _, errStat := os.Stat(nameIndexPath); os.IsNotExist(errStat) {
		log.Printf("Name index not found in %s\nBuilding it...", nameIndexPath)
		nameFinder = name.BuildIndex(cities)
		err = nameFinder.SerializeIndex(nameIndexPath)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize name index: %v", err)
		}
	} else {
		nameFinder, err = name.DeserializeIndex(nameIndexPath)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize name index: %v", err)
		}
	}
	return nameFinder, nil
}

func ensurePostalCodeIndex(cfg *config.Config, postalCodes map[string]map[string]dataLoader.PostalCodeEntry) (*postalCode.Finder, error) {
	postalCodeIndexPath := filepath.Join(cfg.DatasetsFolder, cfg.PostalCodeIndexFile)
	var postalCodeFinder *postalCode.Finder
	var err error

	log.Printf("Ensuring postal code index is built and serialized in %s", postalCodeIndexPath)
	if _, errStat := os.Stat(postalCodeIndexPath); os.IsNotExist(errStat) {
		log.Printf("Postal code index not found in %s\nBuilding it...", postalCodeIndexPath)
		postalCodeFinder = postalCode.BuildIndex(postalCodes)
		err = postalCodeFinder.SerializeIndex(postalCodeIndexPath)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize postal code index: %v", err)
		}
	} else {
		postalCodeFinder, err = postalCode.DeserializeIndex(postalCodeIndexPath)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize postal code index: %v", err)
		}
	}
	return postalCodeFinder, nil
}
