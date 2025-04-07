package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"

	"git.divar.cloud/divar/girls-hackathon/realestate-poi/pkg/database/db"
	"git.divar.cloud/divar/girls-hackathon/realestate-poi/pkg/transport"
	"github.com/jackc/pgx/v5/pgtype"
)

type TransportService struct {
	query            *db.Queries
	priceProviders   map[string]transport.PriceProvider
	locationProvider transport.LocationProvider
}

func NewTransportService(snapp *transport.Snapp, tapsi *transport.Tapsi, neshan *transport.Neshan, query *db.Queries) *TransportService {
	return &TransportService{
		priceProviders: map[string]transport.PriceProvider{
			"snapp": snapp,
			"tapsi": tapsi,
		},
		locationProvider: neshan,
		query:            query,
	}
}

type GetPriceResponse struct {
	SnappPrice int `json:"snapp_price"`
	TapsiPrice int `json:"tapsi_price"`
}

func (t *TransportService) GetPrice(ctx context.Context, originLat, originLong, destinationLat, destinationLong string) (*GetPriceResponse, error) {
	response := &GetPriceResponse{}
	for name, client := range t.priceProviders {
		price, err := client.GetPriceEstimation(ctx, originLat, originLong, destinationLat, destinationLong)
		if err != nil {
			log.Printf("%s price error: %v", name, err)
			continue
		}
		switch name {
		case "snapp":
			response.SnappPrice = price
		case "tapsi":
			response.TapsiPrice = price
		}

	}
	if response.SnappPrice == 0 && response.TapsiPrice == 0 {
		return response, fmt.Errorf("no price available from any provider")
	}
	return response, nil
}

func (s *TransportService) FindNearestStation(ctx context.Context, userId, postToken, originLat, originLong string) (*transport.NearbyPOIsResponse, error) {

	//check if this user is a valid peerson using post purchase, tokens

	// Convert string coordinates to float64
	lat, err := strconv.ParseFloat(originLat, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid latitude: %w", err)
	}

	lng, err := strconv.ParseFloat(originLong, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid longitude: %w", err)
	}

	result, err := s.query.Get3NearestPoiFromEachType(ctx, db.Get3NearestPoiFromEachTypeParams{
		Longitude: lng,
		Latitude:  lat})

	if err == nil && len(result) > 0 {
		log.Printf("Found POIs in database for coordinates (%s, %s)", originLat, originLong)

		response := &transport.NearbyPOIsResponse{
			Subway:      &transport.POICategory{},
			BusStation:  &transport.POICategory{},
			Hospital:    &transport.POICategory{},
			Supermarket: &transport.POICategory{},
			FruitMarket: &transport.POICategory{},
		}

		// Process the results and populate the response
		for _, poi := range result {
			// Add logic here to populate the appropriate category based on poi.Type
			// This is a simplified example - adjust according to your actual data structure
			switch poi.PoiType {
			case "subway":
				response.Subway.POIs = append(response.Subway.POIs, convertToTransportPOI(poi))
			case "bus_station":
				response.BusStation.POIs = append(response.BusStation.POIs, convertToTransportPOI(poi))
			case "hospital":
				response.Hospital.POIs = append(response.Hospital.POIs, convertToTransportPOI(poi))
			case "super_market":
				response.Supermarket.POIs = append(response.Supermarket.POIs, convertToTransportPOI(poi))
			case "fruit_market":
				response.FruitMarket.POIs = append(response.FruitMarket.POIs, convertToTransportPOI(poi))
			}
		}
		return response, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("Database error when checking cache: %v", err)
		return nil, fmt.Errorf("error checking cache: %w", err)
	}

	apiResult, err := s.locationProvider.GetAllNearbyPOIs(ctx, originLat, originLong, 3)
	if err != nil {
		return nil, fmt.Errorf("error calling external API: %w", err)
	}

	go s.cacheAllPOIResults(context.Background(), postToken, apiResult)

	return apiResult, nil

}

func convertToTransportPOI(poi db.Get3NearestPoiFromEachTypeRow) *transport.POIInfo {
	return &transport.POIInfo{
		Name:     poi.PoiName,
		Address:  poi.PoiAddress.String,
		Lat:      poi.PoiLatitude,
		Long:     poi.PoiLongitude,
		Duration: poi.Duration,
		Distance: poi.Distance,
	}
}

func (t *TransportService) cacheAllPOIResults(ctx context.Context, postId string, apiResult *transport.NearbyPOIsResponse) {

	if apiResult.Subway != nil && len(apiResult.Subway.POIs) > 0 {
		t.processPOICategory(ctx, postId, apiResult.Subway.POIs, "subway")
	}

	if apiResult.BusStation != nil && len(apiResult.BusStation.POIs) > 0 {
		t.processPOICategory(ctx, postId, apiResult.BusStation.POIs, "bus_station")
	}

	if apiResult.Hospital != nil && len(apiResult.Hospital.POIs) > 0 {
		t.processPOICategory(ctx, postId, apiResult.Hospital.POIs, "hospital")
	}

	if apiResult.Supermarket != nil && len(apiResult.Supermarket.POIs) > 0 {
		t.processPOICategory(ctx, postId, apiResult.Supermarket.POIs, "super_market")
	}

	if apiResult.FruitMarket != nil && len(apiResult.FruitMarket.POIs) > 0 {
		t.processPOICategory(ctx, postId, apiResult.FruitMarket.POIs, "fruit_market")
	}
}
func (t *TransportService) processPOICategory(ctx context.Context, postId string, pois []*transport.POIInfo, poiType string) {
	for _, poi := range pois {
		// Upsert the POI
		poiId, err := t.query.UpsertPOI(ctx, db.UpsertPOIParams{
			Latitude:  poi.Lat,
			Longitude: poi.Long,
			Name:      poi.Name,
			Address:   pgtype.Text{String: poi.Address, Valid: true},
			Type:      db.PoiType(poiType),
		})

		if err != nil {
			log.Printf("Error upserting %s POI: %v", poiType, err)
			continue
		}

		result, err := t.query.SaveTravelMetrics(ctx, db.SaveTravelMetricsParams{
			Distance:      poi.Distance,
			Duration:      poi.Duration,
			OriginID:      postId,
			DestinationID: poiId,
		})

		if err != nil {
			log.Printf("Error saving travel metrics for %s: %v", poiType, err)
			continue
		}

		if result.RowsAffected() == 0 {
			log.Printf("Travel metrics already exist for origin %s and destination %d (%s)", postId, poiId, poiType)
		} else {
			log.Printf("Successfully cached %s '%s' for post ID %s", poiType, poi.Name, postId)
		}
	}
}
