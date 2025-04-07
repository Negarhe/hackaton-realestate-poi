package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"

	"git.divar.cloud/divar/girls-hackathon/realestate-poi/pkg/configs"
)

const (
	searchApi    = "https://api.neshan.org/v1/search?term=%s&lat=%f&lng=%f"
	directionApi = "https://api.neshan.org/v4/direction/?origin=%s,&destination=%s"
)

type SearchResult struct {
	Items []Items `json:"items"`
}

type DirectionResult struct {
	Route []Routes `json:"routes"`
}

type Routes struct {
	Legs []Leg `json:"legs"`
}

type Leg struct {
	Duration Duration `json:"duration"`
	Distance Distance `json:"distance"`
	Steps    []Step   `json:"steps"`
}

type Step struct {
	Instruction string    `json:"instruction"`
	Distance    *Distance `json:"distance"`
	Duration    *Duration `json:"duration"`
}

type Distance struct {
	Value float64 `json:"value"`
	Text  string  `json:"text"`
}

type Duration struct {
	Value float64 `json:"value"`
	Text  string  `json:"text"`
}

type Items struct {
	Title         string         `json:"title"`
	Address       string         `json:"address"`
	Category      string         `json:"category"`
	Type          string         `json:"type"`
	Region        string         `json:"region"`
	Neighbourhood string         `json:"neighbourhood"`
	Location      SearchLocation `json:"location"`
}
type ItemWithDistance struct {
	item     Items
	distance float64
}
type SearchLocation struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}
type StationResponse struct {
	ClosestStation string  `json:"closest"`
	StationLat     float64 `json:"stationLat"`
	StationLong    float64 `json:"stationLong"`
	TotalDuration  string  `json:"totalDuration"`
	TotalDistance  string  `json:"totalDistance"`
}

type Neshan struct {
	apiKey string
}

type POIInfo struct {
	Name     string  `json:"name"`
	Address  string  `json:"address"`
	Lat      float64 `json:"lat"`
	Long     float64 `json:"long"`
	Duration int32   `json:"duration"`
	Distance int32   `json:"distance"`
}

type POICategory struct {
	POIs []*POIInfo `json:"POIs"`
}

type NearbyPOIsResponse struct {
	Subway      *POICategory `json:"subway,omitempty"`
	BusStation  *POICategory `json:"bus_station,omitempty"`
	Hospital    *POICategory `json:"hospital,omitempty"`
	Supermarket *POICategory `json:"super_market,omitempty"`
	FruitMarket *POICategory `json:"fruit_market,omitempty"`
}

func NewNeshan(s *configs.NeshanConfig) *Neshan {
	return &Neshan{
		apiKey: s.NeshanApiKey,
	}

}

const (
	subwayUrlEncoded      = "%D9%85%D8%AA%D8%B1%D9%88"
	busStationUrlEncoded  = "%D8%A7%DB%8C%D8%B3%D8%AA%DA%AF%D8%A7%D9%87%20%D8%A7%D8%AA%D9%88%D8%A8%D9%88%D8%B3"
	hospitalUrlEncoded    = "%D8%A8%DB%8C%D9%85%D8%A7%D8%B1%D8%B3%D8%AA%D8%A7%D9%86%0A"
	supermarketUrlEncoded = "%D8%B3%D9%88%D9%BE%D8%B1%D9%85%D8%A7%D8%B1%DA%A9%D8%AA"
	fruitMarketUrlEncoded = "%D8%A8%D8%A7%D8%B2%D8%A7%D8%B1%20%D9%85%DB%8C%D9%88%D9%87"
)

// func (n *Neshan) GetSearchResult(startLat, startLong float64) ([]Items, error) {
// 	searchURL := fmt.Sprintf(searchApi, "%D9%85%D8%AA%D8%B1%D9%88", startLat, startLong)
// 	req, err := http.NewRequest("GET", searchURL, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	req.Header.Set("Api-Key", n.apiKey)
// 	client := &http.Client{}

// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("error: %s", resp.Status)
// 	}

// 	var searchResult SearchResult
// 	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
// 		return nil, err
// 	}

// 	if len(searchResult.Items) == 0 {
// 		return nil, fmt.Errorf("no station found")
// 	}

// 	var nearbyStations []Items
// 	for _, item := range searchResult.Items {
// 		if distance(startLat, startLong, item.Location.Y, item.Location.X) <= 4 {
// 			nearbyStations = append(nearbyStations, item)
// 		}
// 	}

// 	if len(nearbyStations) == 0 {
// 		return nil, fmt.Errorf("no station found within 4 km radius")
// 	}

// 	return nearbyStations, nil
// }

func (n *Neshan) GetDirectionResult(origin, destination string) (DirectionResult, error) {
	directionUrl := fmt.Sprintf(directionApi, origin, destination)
	req, err := http.NewRequest("GET", directionUrl, nil)
	if err != nil {
		return DirectionResult{}, err
	}

	req.Header.Set("Api-Key", n.apiKey)
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return DirectionResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return DirectionResult{}, fmt.Errorf("error: %s", resp.Status)
	}

	var directionResult DirectionResult
	if err := json.NewDecoder(resp.Body).Decode(&directionResult); err != nil {
		return DirectionResult{}, err
	}

	return directionResult, nil
}

func distance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Radius of the Earth in km
	dLat := (lat2 - lat1) * (math.Pi / 180)
	dLon := (lon2 - lon1) * (math.Pi / 180)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1*(math.Pi/180))*math.Cos(lat2*(math.Pi/180))*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func PersianToEnglishNumerals(input string) string {
	replacements := map[rune]rune{
		'۰': '0',
		'۱': '1',
		'۲': '2',
		'۳': '3',
		'۴': '4',
		'۵': '5',
		'۶': '6',
		'۷': '7',
		'۸': '8',
		'۹': '9',
	}
	result := []rune(input)
	for i, char := range result {
		if replacement, found := replacements[char]; found {
			result[i] = replacement
		}
	}
	return string(result)
}

func convertToMeters(distance, unit string) (string, error) {
	value, err := strconv.ParseFloat(distance, 64)
	if err != nil {
		return "", err
	}
	switch unit {
	case "متر":
		return fmt.Sprintf("%.0f", value), nil
	case "کیلومتر":
		return fmt.Sprintf("%.0f", value*1000), nil
	default:
		return "", fmt.Errorf("unknown unit: %s", unit)
	}
}
func (n *Neshan) GetSearchResult(startLat, startLong float64, poiType string) ([]*ItemWithDistance, error) {
	searchURL := fmt.Sprintf(searchApi, poiType, startLat, startLong)
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Api-Key", n.apiKey)
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: %s", resp.Status)
	}

	var searchResult SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return nil, err
	}

	if len(searchResult.Items) == 0 {
		return nil, fmt.Errorf("no POI found")
	}

	var nearbyPOIs []*ItemWithDistance
	dist := 0.0
	// Filter POIs within 4 km radius
	for _, item := range searchResult.Items {
		dist = distance(startLat, startLong, item.Location.Y, item.Location.X)
		if dist <= 4 {
			nearbyPOIs = append(nearbyPOIs, &ItemWithDistance{item: item, distance: dist})
		}
	}
	sort.Slice(nearbyPOIs, func(i, j int) bool {
		return nearbyPOIs[i].distance < nearbyPOIs[j].distance
	})

	if len(nearbyPOIs) == 0 {
		return nil, fmt.Errorf("no POI found within 4 km radius")
	}

	return nearbyPOIs, nil
}

// func (n *Neshan) GetSubwayStation(startLatstr, startLongstr string) (*StationResponse, error) {
// 	if startLatstr == "" || startLongstr == "" {
// 		return nil, fmt.Errorf("startLat and startLong are required")
// 	}
// 	startLat, err := strconv.ParseFloat(startLatstr, 64)
// 	if err != nil {
// 		return nil, fmt.Errorf("invalid startLat parameter")
// 	}

// 	startLong, err := strconv.ParseFloat(startLongstr, 64)
// 	if err != nil {
// 		return nil, fmt.Errorf("invalid startLong parameter")
// 	}

// 	nearbyStations, err := n.GetSearchResult(startLat, startLong)
// 	if err != nil {
// 		return nil, err
// 	}
// 	log.Println(nearbyStations)

// 	closest := nearbyStations[1]
// 	stationLat := closest.Location.Y
// 	stationLong := closest.Location.X
// 	origin := fmt.Sprintf("%f", startLat) + "," + fmt.Sprintf("%f", startLong)
// 	destination := fmt.Sprintf("%f", closest.Location.Y) + "," + fmt.Sprintf("%f", closest.Location.X)

// 	directionResult, err := n.GetDirectionResult(origin, destination)
// 	if err != nil {
// 		return nil, err
// 	}

// 	totalDuration := PersianToEnglishNumerals(strings.Split(directionResult.Route[0].Legs[0].Duration.Text, " ")[0])

// 	distanceParts := strings.Split(directionResult.Route[0].Legs[0].Distance.Text, " ")
// 	distanceValue := PersianToEnglishNumerals(distanceParts[0])
// 	distanceUnit := distanceParts[1]

// 	totalDistance, err := convertToMeters(distanceValue, distanceUnit)
// 	if err != nil {
// 		return nil, fmt.Errorf("error converting distance: %v", err)
// 	}

// 	response := &StationResponse{
// 		ClosestStation: closest.Title,
// 		StationLat:     stationLat,
// 		StationLong:    stationLong,
// 		TotalDuration:  totalDuration,
// 		TotalDistance:  totalDistance,
// 	}
// 	return response, nil
// }

func (n *Neshan) findNearestPOI(startLat, startLong float64, poiType string, limit int) (*POICategory, error) {
	// Get nearby POIs
	nearbyPOIs, err := n.GetSearchResult(startLat, startLong, poiType)
	if err != nil || len(nearbyPOIs) == 0 {
		return nil, err
	}

	// Find the truly closest POI
	// var closest Items
	// var minDistance float64 = math.MaxFloat64

	if len(nearbyPOIs) > limit {
		nearbyPOIs = nearbyPOIs[:limit]
	}
	result := POICategory{
		POIs: make([]*POIInfo, 0, len(nearbyPOIs)),
	}

	for _, poi := range nearbyPOIs {

		// Get directions
		origin := fmt.Sprintf("%f,%f", startLat, startLong)
		destination := fmt.Sprintf("%f,%f", poi.item.Location.Y, poi.item.Location.X)

		directionResult, err := n.GetDirectionResult(origin, destination)
		if err != nil {
			return nil, err
		}

		if len(directionResult.Route) == 0 || len(directionResult.Route[0].Legs) == 0 {
			return nil, fmt.Errorf("no route found")
		}

		leg := directionResult.Route[0].Legs[0]

		// Process duration
		totalDuration, err := strconv.Atoi(PersianToEnglishNumerals(strings.Split(leg.Duration.Text, " ")[0]))
		if err != nil {
			return nil, fmt.Errorf("error converting duration to int: %v", err)
		}

		// Process distance
		distanceParts := strings.Split(leg.Distance.Text, " ")
		distanceValue := PersianToEnglishNumerals(distanceParts[0])
		distanceUnit := distanceParts[1]

		totalDistance, err := convertToMeters(distanceValue, distanceUnit)
		if err != nil {
			return nil, fmt.Errorf("error converting distance: %v", err)
		}

		distanceVal, _ := strconv.Atoi(totalDistance)
		result.POIs = append(result.POIs, &POIInfo{
			Name:     poi.item.Title,
			Address:  poi.item.Address,
			Lat:      poi.item.Location.Y,
			Long:     poi.item.Location.X,
			Duration: int32(totalDuration),
			Distance: int32(distanceVal),
		})
	}

	return &result, nil
}

// Main function to get all nearby POIs
func (n *Neshan) GetAllNearbyPOIs(ctx context.Context, startLatstr, startLongstr string, limit int) (*NearbyPOIsResponse, error) {
	// Validate input parameters
	if startLatstr == "" || startLongstr == "" {
		return nil, fmt.Errorf("startLat and startLong are required")
	}

	startLat, err := strconv.ParseFloat(startLatstr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid startLat parameter")
	}

	startLong, err := strconv.ParseFloat(startLongstr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid startLong parameter")
	}

	var wg sync.WaitGroup
	response := &NearbyPOIsResponse{}

	poiResults := make(map[string]*POICategory)
	var resultMutex sync.Mutex // Mutex to protect the map

	// Define POI types to search for
	poiTypes := map[string]string{
		"subway":       subwayUrlEncoded,
		"bus_station":  busStationUrlEncoded,
		"hospital":     hospitalUrlEncoded,
		"super_market": supermarketUrlEncoded,
		"fruit_market": fruitMarketUrlEncoded,
	}

	// Process each POI type in parallel
	for poiKey, poiEncoded := range poiTypes {
		wg.Add(1)
		go func(key, encodedType string) {
			defer wg.Done()

			poi, err := n.findNearestPOI(startLat, startLong, encodedType, limit)
			if err != nil {
				log.Printf("Error finding %s: %v", key, err)
				return
			}

			// Safely store the result
			resultMutex.Lock()
			poiResults[key] = poi
			resultMutex.Unlock()
		}(poiKey, poiEncoded)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Assign results to response struct
	response.Subway = poiResults["subway"]
	response.BusStation = poiResults["bus_station"]
	response.Hospital = poiResults["hospital"]
	response.Supermarket = poiResults["super_market"]
	response.FruitMarket = poiResults["fruit_market"]

	return response, nil
}
