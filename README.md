# Nearest City Finder for Go

A high-performance Go library to find the nearest city based on geographical coordinates using the S2 Geometry Library.

## Features

- **Efficient Nearest City Search**: Uses the S2 Geometry Library to provide fast and accurate nearest city searches.
- **Low Memory Consumption**: Optimized for low memory usage compared to other data structures.
- **Easy Integration**: Simple API for integrating into your Go projects.
- **Support for Postal Codes**: Find cities based on postal codes.

## Why S2?

The S2 Geometry Library is chosen for its superior performance and efficiency in handling geographical data. It provides:
- **Hierarchical Spatial Indexing**: Efficiently manages and queries large sets of geographical points.
- **High Precision**: Ensures accurate results even with large datasets.
- **Low Memory Footprint**: Uses memory efficiently, making it suitable for applications with limited resources.

## Performance Comparison

In benchmark tests, the S2 Geometry Library outperformed other methods such as R-tree, k-d Tree, and Geohash in both speed and memory consumption. Below is a summary of the performance comparison based on a sample dataset of approximately 12.8 million cities.

### Dataset Sizes

- **All Countries Dataset**: 12,759,551 entries
- **Postal Code Dataset**: 100,000+ entries (example size)

### Speed Comparison (in milliseconds)

| City         | Finder   | Time (ms)         | Memory (KB)       | Nearest City                        | Latitude    | Longitude    |
|--------------|----------|-------------------|-------------------|-------------------------------------|-------------|--------------|
| London       | k-d Tree | **1 <- Fastest**  | 648               | Nelson's Column                     | 51.507770   | -0.127920    |
| London       | R-tree   | 89                | 624               | St Martin-in-the-Fields             | 51.508860   | -0.126900    |
| London       | S2       | **0 <- Lowest**   | **6 <- Lowest**   | Edith Cavell Memorial               | 51.509390   | -0.127160    |
| London       | Geohash  | 3319              | 238               | Nelson's Column                     | 51.507770   | -0.127920    |
| Tokyo        | k-d Tree | 7                 | 649               | Tokyo Prefecture                    | 35.689500   | 139.691710   |
| Tokyo        | R-tree   | 26                | 606               | Tokyo                               | 35.689500   | 139.691710   |
| Tokyo        | S2       | **4 <- Lowest**   | 362               | Tōkyō Tochōsha                      | 35.689440   | 139.691780   |
| Tokyo        | Geohash  | 2987              | 238               | Tokyo Prefecture                    | 35.689500   | 139.691710   |
| Sydney       | k-d Tree | 2                 | 454               | Royal Theatre                       | -33.868000  | 151.209000   |
| Sydney       | R-tree   | 20                | 409               | Medina Grand Harbourside            | -33.867600  | 151.209600   |
| Sydney       | S2       | **0 <- Lowest**   | **8 <- Lowest**   | Westin Sydney Heritage Superior     | -33.867790  | 151.207750   |
| Sydney       | Geohash  | 3114              | 65                | Royal Theatre                       | -33.868000  | 151.209000   |
| Moscow       | k-d Tree | 4                 | 682               | Budappest                           | 55.755790   | 37.617630    |
| Moscow       | R-tree   | 36                | 537               | Bolschoi-Theater                    | 55.760300   | 37.618600    |
| Moscow       | S2       | **1 <- Lowest**   | **5 <- Lowest**   | Budappest                           | 55.755790   | 37.617630    |
| Moscow       | Geohash  | 3325              | 238               | Budappest                           | 55.755790   | 37.617630    |
| Chicago      | k-d Tree | 2                 | 453               | Kluczynski Federal Building         | 41.878370   | -87.630050   |
| Chicago      | S2       | **0 <- Lowest**   | **10 <- Lowest**  | HHS Region 5                        | 41.877860   | -87.629520   |
| Chicago      | R-tree   | 222               | 629               | Federal Center                      | 41.878920   | -87.629770   |
| Chicago      | Geohash  | 3167              | 238               | Kluczynski Federal Building         | 41.878370   | -87.630050   |
| New York     | k-d Tree | 2                 | 451               | New York City Hall                  | 40.712600   | -74.005970   |
| New York     | S2       | **0 <- Lowest**   | **11 <- Lowest**  | Vanderlyn's Rotunda (historical)    | 40.713160   | -74.004310   |
| New York     | R-tree   | 218               | 651               | Old New York County Courthouse      | 40.713440   | -74.005420   |
| New York     | Geohash  | 3326              | 239               | New York City Hall                  | 40.712600   | -74.005970   |
| Paris        | k-d Tree | **0 <- Lowest**   | 644               | Hôtel de Ville de Paris             | 48.856440   | 2.352440     |
| Paris        | S2       | 5                 | **8 <- Lowest**   | Rue de la Coutellerie               | 48.857350   | 2.350270     |
| Paris        | R-tree   | 201               | 611               | Rue de la Verrerie                  | 48.857830   | 2.353350     |
| Paris        | Geohash  | 3176              | 61                | Hôtel de Ville de Paris             | 48.856440   | 2.352440     |

### Memory Consumption (KB)

| Finder   | Memory (KB)  |
|----------|--------------|
| S2       | **8,369,107 <- Lowest**  |
| Geohash  | 9,130,174    |
| k-d Tree | 13,826,040   |
| R-tree   | 20,418,544   |

### Conclusion

The S2 Geometry Library offers the best performance in terms of both speed and memory consumption for finding the nearest city based on geographical coordinates. This makes it the ideal choice for applications that require efficient and accurate spatial queries.

## Installation

To install the library, use `go get`:


```bash
go get github.com/SamyRai/nearestcity
```

## Usage

### Finding the Nearest City

To find the nearest city based on latitude and longitude:

```go
package main

import (
    "fmt"
    "log"

    "github.com/SamyRai/cityFinder"
)

func main() {
	
    // Find the nearest city
    nearestCity := cityFinder.FindNearestCity(40.7128, -74.0060) // New York coordinates

    if nearestCity != nil {
        fmt.Printf("Nearest city: %s, %s\n", nearestCity.Name, nearestCity.Country)
    } else {
        log.Println("No city found")
    }
}
```

### Finding a City by Postal Code

To find a city based on postal code:

```go
nearestCity := cityFinder.FindCityByPostalCode("10001") // New York postal code

if nearestCity != nil {
    fmt.Printf("Nearest city: %s, %s\n", nearestCity.Name, nearestCity.Country)
} else {
    log.Println("No city found")
}
```

## Running the Server

The package also includes a server that provides an API for finding the nearest city and querying cities by name or postal code.

### Building the Server

To build the server, use the following command:

```bash
go build -o nearestcityserver cmd/server/main.go
```

### Running the Server

To run the server, execute the built binary:

```bash
./nearestcityserver
```

By default, the server will listen on port 3000.

### API Endpoints

- **Find Nearest City**: `/nearest?lat=<latitude>&lon=<longitude>`
- **Find City by Name**: `/coordinates?name=<city_name>`
- **Find City by Postal Code**: `/postalcode?postalcode=<postal_code>`

## Testing

Unit tests are included for the library. To run the tests, use the following command:

```bash
go test ./cmd/server
```

## Initialization and Datasets

This project requires datasets from the [GeoNames](http://www.geonames.org/) database. Specifically, you need the `allCountries.txt` for city data and `zipCodes.txt` for postal code data. These files should be placed in the `datasets` folder.

During initialization, the application checks if these datasets and the S2 index are available. If they are not, it downloads and extracts the required datasets and builds the S2 index. This ensures that the necessary data is always available regardless of how the library is used.

### Using the Server

The server can be started using the following command:

```bash
go run cmd/server/main.go
```

## Contributing

Contributions are welcome! Please fork the repository and submit pull requests for any improvements or bug fixes.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Acknowledgements

- [S2 Geometry Library](https://github.com/golang/geo) for providing the efficient spatial indexing and search capabilities.
- All contributors and community members for their support and contributions.
- [GeoNames](http://www.geonames.org/) for providing the geographical data used in this project.