package rtree

import (
	"cityFinder/city"
	"encoding/gob"
	"os"

	"github.com/cheggaaa/pb/v3"
	"github.com/dhconnelly/rtreego"
)

type SerializableRTree struct {
	Entries []city.SpatialCity
}

func BuildRTree(cities []city.SpatialCity) *rtreego.Rtree {
	rtree := rtreego.NewTree(2, 25, 50)
	bar := pb.Full.Start(len(cities))
	defer bar.Finish()

	for _, city := range cities {
		rtree.Insert(&city)
		bar.Increment()
	}
	return rtree
}

func FindNearestCity(lat, lon float64, rtree *rtreego.Rtree) *city.City {
	point := rtreego.Point{lon, lat}
	rect, _ := rtreego.NewRect(point, []float64{0.00001, 0.00001})
	results := rtree.SearchIntersect(rect)

	bar := pb.Full.Start(len(results))
	defer bar.Finish()

	minDistance := float64(1<<63 - 1)
	var nearestCity *city.City

	for _, item := range results {
		spatialCity := item.(*city.SpatialCity)
		spatialCityAsPoint := rtreego.Point{spatialCity.Longitude, spatialCity.Latitude}
		distance := city.EuclideanDistance(point, spatialCityAsPoint)
		if distance < minDistance {
			minDistance = distance
			nearestCity = &spatialCity.City
		}
		bar.Increment()
	}
	return nearestCity
}

func SaveRTree(filepath string, rtree *rtreego.Rtree) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	var entries []city.SpatialCity
	bb, _ := rtreego.NewRect(rtreego.Point{-180, -90}, []float64{360, 180}) // bounding box covering the whole world
	results := rtree.SearchIntersect(bb)

	bar := pb.Full.Start(len(results))
	defer bar.Finish()

	for _, item := range results {
		entries = append(entries, *item.(*city.SpatialCity))
		bar.Increment()
	}

	serializableRTree := SerializableRTree{
		Entries: entries,
	}

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(serializableRTree); err != nil {
		return err
	}
	return nil
}

func LoadRTree(filepath string) (*rtreego.Rtree, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var serializableRTree SerializableRTree
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&serializableRTree); err != nil {
		return nil, err
	}

	rtree := rtreego.NewTree(2, 25, 50)

	bar := pb.Full.Start(len(serializableRTree.Entries))
	defer bar.Finish()

	for _, entry := range serializableRTree.Entries {
		rtree.Insert(&entry)
		bar.Increment()
	}
	return rtree, nil
}
