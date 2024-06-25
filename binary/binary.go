package binary

import (
	"cityFinder/city"
	"encoding/binary"
	"os"

	"github.com/cheggaaa/pb/v3"
	"github.com/dhconnelly/rtreego"
)

func LoadBinary(filepath string) ([]city.SpatialCity, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := fi.Size()

	var cities []city.SpatialCity
	bar := pb.Full.Start64(fileSize)
	defer bar.Finish()

	for bytesRead := int64(0); bytesRead < fileSize; {
		var lat, lon float64
		var nameLen, countryLen int32

		if err := binary.Read(file, binary.LittleEndian, &lat); err != nil {
			return nil, err
		}
		bytesRead += 8

		if err := binary.Read(file, binary.LittleEndian, &lon); err != nil {
			return nil, err
		}
		bytesRead += 8

		if err := binary.Read(file, binary.LittleEndian, &nameLen); err != nil {
			return nil, err
		}
		bytesRead += 4

		name := make([]byte, nameLen)
		if _, err := file.Read(name); err != nil {
			return nil, err
		}
		bytesRead += int64(nameLen)

		if err := binary.Read(file, binary.LittleEndian, &countryLen); err != nil {
			return nil, err
		}
		bytesRead += 4

		country := make([]byte, countryLen)
		if _, err := file.Read(country); err != nil {
			return nil, err
		}
		bytesRead += int64(countryLen)

		cityObj := city.City{
			Latitude:  lat,
			Longitude: lon,
			Name:      string(name),
			Country:   string(country),
		}

		point := rtreego.Point{lon, lat}
		rect, _ := rtreego.NewRect(point, []float64{0.00001, 0.00001})
		spatialCity := city.SpatialCity{City: cityObj, Rect: rect}

		cities = append(cities, spatialCity)
		bar.SetCurrent(bytesRead)
	}
	return cities, nil
}

func SaveBinary(filepath string, cities []city.SpatialCity) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	bar := pb.Full.Start(len(cities))
	defer bar.Finish()

	for _, city := range cities {
		if err := binary.Write(file, binary.LittleEndian, city.Latitude); err != nil {
			return err
		}
		if err := binary.Write(file, binary.LittleEndian, city.Longitude); err != nil {
			return err
		}
		nameLen := int32(len(city.Name))
		if err := binary.Write(file, binary.LittleEndian, nameLen); err != nil {
			return err
		}
		if _, err := file.Write([]byte(city.Name)); err != nil {
			return err
		}
		countryLen := int32(len(city.Country))
		if err := binary.Write(file, binary.LittleEndian, countryLen); err != nil {
			return err
		}
		if _, err := file.Write([]byte(city.Country)); err != nil {
			return err
		}
		bar.Increment()
	}
	return nil
}
