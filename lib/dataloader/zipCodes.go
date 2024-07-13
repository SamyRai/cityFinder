package dataloader

import (
	"encoding/csv"
	"os"
	"strconv"
)

type PostalCodeEntry struct {
	CountryCode string
	PostalCode  string
	PlaceName   string
	AdminName1  string
	AdminCode1  string
	AdminName2  string
	AdminCode2  string
	AdminName3  string
	AdminCode3  string
	Latitude    float64
	Longitude   float64
	Accuracy    int
}

func LoadPostalCodes(filepath string) (map[string]PostalCodeEntry, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '\t'
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	postalCodes := make(map[string]PostalCodeEntry)
	for _, record := range records {
		lat, _ := strconv.ParseFloat(record[9], 64)
		lon, _ := strconv.ParseFloat(record[10], 64)
		accuracy, _ := strconv.Atoi(record[11])

		postalCode := PostalCodeEntry{
			CountryCode: record[0],
			PostalCode:  record[1],
			PlaceName:   record[2],
			AdminName1:  record[3],
			AdminCode1:  record[4],
			AdminName2:  record[5],
			AdminCode2:  record[6],
			AdminName3:  record[7],
			AdminCode3:  record[8],
			Latitude:    lat,
			Longitude:   lon,
			Accuracy:    accuracy,
		}

		postalCodes[postalCode.PostalCode] = postalCode
	}

	return postalCodes, nil
}
