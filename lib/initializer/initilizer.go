package initializer

import (
	"archive/zip"
	"fmt"
	"github.com/SamyRai/cityFinder/lib/config"
	"github.com/SamyRai/cityFinder/lib/dataloader"
	"github.com/SamyRai/cityFinder/lib/finder"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func Initialize(cfg *config.Config) (*finder.S2Finder, error) {
	if err := ensureDatasets(cfg); err != nil {
		return nil, err
	}
	return ensureS2Index(cfg)
}

func ensureDatasets(cfg *config.Config) error {
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
			os.MkdirAll(fpath, os.ModePerm)
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
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func ensureS2Index(cfg *config.Config) (*finder.S2Finder, error) {
	if _, err := os.Stat(cfg.S2IndexFile); os.IsNotExist(err) {
		log.Printf("S2 index not found, building...")
		if err := buildS2Index(cfg); err != nil {
			return nil, err
		}
	}
	return finder.DeserializeIndex(cfg.S2IndexFile)
}

func buildS2Index(cfg *config.Config) error {
	csvLocation := filepath.Join(cfg.DatasetsFolder, cfg.AllCitiesFile)
	postalCodeLocation := filepath.Join(cfg.DatasetsFolder, cfg.PostalCodesFile)

	log.Printf("Loading the CSV data from %s", csvLocation)
	cities, err := dataloader.LoadGeoNamesCSV(csvLocation)
	if err != nil {
		return fmt.Errorf("failed to load GeoNames data from CSV: %v", err)
	}
	log.Printf("Finished loading CSV data")

	log.Printf("Loading the Postal Code data from %s", postalCodeLocation)
	postalCodes, err := dataloader.LoadPostalCodes(postalCodeLocation)
	if err != nil {
		return fmt.Errorf("failed to load Postal Code data: %v", err)
	}
	log.Printf("Finished loading Postal Code data")

	log.Printf("Building S2 index")
	s2Finder := finder.BuildS2Index(cities, postalCodes)

	log.Printf("Serializing S2 index to %s", cfg.S2IndexFile)
	err = s2Finder.SerializeIndex(cfg.S2IndexFile)
	if err != nil {
		return fmt.Errorf("failed to serialize S2 index: %v", err)
	}
	log.Printf("Finished serializing S2 index")
	return nil
}
