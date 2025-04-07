-- name: SaveTravelMetrics :execresult
INSERT INTO travel_metrics (
    distance, 
    duration, 
    origin_id, 
    destination_id
) VALUES (
    $1, $2, $3, $4
) ON CONFLICT (origin_id, destination_id) 
DO UPDATE SET 
    distance = $1, 
    duration = $2, 
    updated_at = CURRENT_TIMESTAMP;