package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"git.divar.cloud/divar/girls-hackathon/realestate-poi/pkg/configs"
)

type snappPoint struct {
	Lat string `json:"lat"`
	Lng string `json:"lng"`
}

type snappRequest struct {
	Points         []*snappPoint `json:"points"`
	VoucherCode    *string       `json:"voucher_code"`
	ServiceTypes   []int         `json:"service_types"`
	Priceriderecom bool          `json:"priceriderecom"`
	Tag            string        `json:"tag"`
	HurryRaised    int           `json:"hurryRaised"`
}

type snappPrices struct {
	Final int `json:"final"`
}

type snappData struct {
	Prices []*snappPrices `json:"prices"`
}
type snappResponse struct {
	Data *snappData `json:"data"`
}

type Snapp struct {
	accessToken    string
	cookiesession1 string
	clck           string
	ga_Y4QV007ERR  string
	ga             string
	ym_d           string
	ym_uid         string
	ym_isad        string
	clsk           string
}

func NewSnapp(s *configs.SnappConfig) *Snapp {
	return &Snapp{
		accessToken:    s.ApiKey,
		cookiesession1: s.CookieSession,
		clck:           s.Clck,
		clsk:           s.Clsk,
		ym_d:           s.YandexDate,
		ym_uid:         s.YandexUID,
		ym_isad:        s.YandexAd,
		ga_Y4QV007ERR:  s.GATracking,
		ga:             s.GA,
	}

}

func (s *Snapp) GetPriceEstimation(ctx context.Context, originLat, originLong, destinationLat, destinationLong string) (int, error) {
	data := snappRequest{
		Points: []*snappPoint{
			{Lat: originLat, Lng: originLong},
			{Lat: destinationLat, Lng: destinationLong},
		},
		VoucherCode:    nil,
		ServiceTypes:   []int{1, 2, 24},
		Priceriderecom: false,
		Tag:            "0",
		HurryRaised:    0,
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling snapp request data: %v", err)
		return 0, fmt.Errorf("failed to prepare snapp request : %w", err)
	}
	body := bytes.NewReader(dataBytes)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://app.snapp.taxi/api/api-base/v2/passenger/newprice/s/6/0", body)
	if err != nil {
		log.Printf("Error creating snapp request: %v", err)
		return 0, fmt.Errorf("failed to create snapp request: %w", err)

	}
	s.SetHeader(req)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making API call to Snapp: %v", err)
		return 0, fmt.Errorf("failed to connect to Snapp service: %w", err)
	}
	defer resp.Body.Close()

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading Snapp response body: %v", err)
		return 0, fmt.Errorf("failed to read Snapp response: %w", err)
	}

	if resp.StatusCode != 200 {
		log.Printf("Unexpected status code from Snapp API: %d", resp.StatusCode)
		return 0, fmt.Errorf("unexpected response from Snapp service (code: %d)", resp.StatusCode)
	}

	var jsonData snappResponse
	err = json.Unmarshal(bodyText, &jsonData)
	if err != nil {
		log.Printf("Error unmarshaling Snapp response: %v", err)
		return 0, fmt.Errorf("failed to parse Snapp response: %w", err)
	}

	if jsonData.Data == nil || len(jsonData.Data.Prices) == 0 {
		log.Printf("No prices found in Snapp response")
		return 0, fmt.Errorf("no price options available for this Snapp route")
	}

	if jsonData.Data.Prices[0].Final == 0 {
		log.Printf("Received zero price from Snapp")
		return 0, fmt.Errorf("invalid price information (zero price) from Snapp")
	}

	return jsonData.Data.Prices[0].Final / 10, nil
}

func (s *Snapp) SetHeader(req *http.Request) {

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", "https://app.snapp.taxi/pre-ride?rideFrom={%22options%22:{%22serviceType%22:1,%22recommender%22:%22cab%22}}")

	req.Header.Set("Origin", "https://app.snapp.taxi")
	req.Header.Set("Authorization", "Bearer "+s.accessToken)
	req.Header.Set("Cookie", "cookiesession1="+s.cookiesession1+
		"_clck="+s.clck+"_ga_Y4QV007ERR="+s.ga_Y4QV007ERR+"_ga="+s.ga+
		"_ym_uid="+s.ym_uid+"_ym_d="+s.ym_d+"_ym_isad="+s.ym_isad+
		"_clsk="+s.clsk)

}
