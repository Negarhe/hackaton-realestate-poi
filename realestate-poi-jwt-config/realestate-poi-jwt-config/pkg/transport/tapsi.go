package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"git.divar.cloud/divar/girls-hackathon/realestate-poi/pkg/configs"
	"github.com/andybalholm/brotli"
)

type origin struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type destinations struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type metadata struct {
	FlowType    string `json:"flowType"`
	PreviewType string `json:"previewType"`
}

type tapsiRequest struct {
	Origin       *origin         `json:"origin"`
	Destinations []*destinations `json:"destinations"`
	Rider        *string         `json:"rider"`
	HasReturn    bool            `json:"hasReturn"`
	WaitingTime  int             `json:"waitingTime"`
	Gateway      string          `json:"gateway"`
	InitiatedVia string          `json:"initiatedVia"`
	Metadata     *metadata       `json:"metadata"`
}

type tapsiPrices struct {
	PassengerShare int `json:"passengerShare"`
}

type tapsiService struct {
	Prices []*tapsiPrices `json:"prices"`
}

type tapsiItems struct {
	Service *tapsiService `json:"service"`
}

type tapsiCategories struct {
	Items []*tapsiItems `json:"items"`
}

type tapsiData struct {
	Categories []*tapsiCategories `json:"categories"`
}

type tapsiResponse struct {
	Data *tapsiData `json:"data"`
}

type Tapsi struct {
	clck         string
	accessToken  string
	refreshToken string
	clsk         string
}

func NewTapsi(s *configs.TapsiConfig) *Tapsi {
	return &Tapsi{
		refreshToken: s.RefreshToken,
		accessToken:  s.AccessToken,
		clck:         s.Clck,
		clsk:         s.Clsk,
	}

}

func (t *Tapsi) GetPriceEstimation(ctx context.Context, stroriginLat, stroriginLong, strdestinationLat, strdestinationLong string) (int, error) {

	originLat, err := strconv.ParseFloat(stroriginLat, 64)
	if err != nil {
		log.Printf("Error parsing origin latitude: %v", err)
		return 0, fmt.Errorf("invalid origin latitude format: %w", err)
	}

	originLong, err := strconv.ParseFloat(stroriginLong, 64)
	if err != nil {
		log.Printf("Error parsing origin longitude: %v", err)
		return 0, fmt.Errorf("invalid origin longitude format: %w", err)
	}

	destinationLat, err := strconv.ParseFloat(strdestinationLat, 64)
	if err != nil {
		log.Printf("Error parsing destination latitude: %v", err)
		return 0, fmt.Errorf("invalid destination latitude format: %w", err)
	}

	destinationLong, err := strconv.ParseFloat(strdestinationLong, 64)
	if err != nil {
		log.Printf("Error parsing destination longitude: %v", err)
		return 0, fmt.Errorf("invalid destination longitude format: %w", err)
	}

	data := tapsiRequest{
		Origin: &origin{
			Latitude:  originLat,
			Longitude: originLong,
		},
		Destinations: []*destinations{
			{
				Latitude:  destinationLat,
				Longitude: destinationLong,
			},
		},
		Rider:        nil,
		HasReturn:    false,
		WaitingTime:  0,
		Gateway:      "CAB",
		InitiatedVia: "WEB",
		Metadata: &metadata{
			FlowType:    "DESTINATION_FIRST",
			PreviewType: "ORIGIN_FIRST",
		},
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling tapsi request data: %v", err)
		return 0, fmt.Errorf("failed to prepare tapsi request: %w", err)
	}
	body := bytes.NewReader(dataBytes)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.tapsi.cab/api/v3/ride/preview", body)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	t.SetHeader(req)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making API call to Tapsi: %v", err)
		return 0, fmt.Errorf("failed to connect to Tapsi service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("Unexpected status code from Tapsi API: %d", resp.StatusCode)
		return 0, fmt.Errorf("unexpected response from Tapsi service (code: %d)", resp.StatusCode)
	}
	reader := brotli.NewReader(resp.Body)
	bodyText, err := io.ReadAll(reader)
	if err != nil {
		log.Printf("Error reading tapsi decompressed response: %v", err)
		return 0, fmt.Errorf("failed to read tapsi response: %w", err)
	}
	var jsonData tapsiResponse
	err = json.Unmarshal(bodyText, &jsonData)
	if err != nil {
		log.Printf("Error unmarshaling response: %v", err)
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	if jsonData.Data == nil {
		log.Printf("No data found in Tapsi response")
		return 0, fmt.Errorf("no price data available from Tapsi")
	}

	if len(jsonData.Data.Categories) == 0 {
		log.Printf("No categories found in Tapsi response")
		return 0, fmt.Errorf("no service categories available for this route")
	}

	if len(jsonData.Data.Categories[0].Items) == 0 {
		log.Printf("No items found in first category")
		return 0, fmt.Errorf("no service options available in the selected category")
	}

	if jsonData.Data.Categories[0].Items[0].Service == nil {
		log.Printf("No service information found in Tapsi response")
		return 0, fmt.Errorf("service details not available")
	}

	if len(jsonData.Data.Categories[0].Items[0].Service.Prices) == 0 {
		log.Printf("No prices found in service")
		return 0, fmt.Errorf("price information not available for this service")
	}

	price := jsonData.Data.Categories[0].Items[0].Service.Prices[0].PassengerShare
	if price == 0 {
		log.Printf("Received zero price from Tapsi")
		return 0, fmt.Errorf("invalid price information (zero price)")
	}

	return price, nil
}

func (t *Tapsi) SetHeader(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:136.0) Gecko/20100101 Firefox/136.0")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", "https://app.tapsi.cab/")
	req.Header.Set("x-agent", "v2.2|passenger|WEBAPP|7.14.7||5.0")
	req.Header.Set("Origin", "https://app.tapsi.cab/")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Priority", "u=4")
	req.Header.Set("TE", "trailers")

	req.Header.Set("Cookie", "_clck="+t.clck+"; "+
		"accessToken="+t.accessToken+"; "+
		"refreshToken="+t.refreshToken+"; "+
		"_clsk="+t.clsk)
}
