package transport

import "context"

type PriceProvider interface {
	GetPriceEstimation(ctx context.Context, originLat, originLong, destinationLat, destinationLong string) (int, error)
}

type LocationProvider interface {
	GetAllNearbyPOIs(ctx context.Context, startLatstr, startLongstr string, limit int) (*NearbyPOIsResponse, error)
}
