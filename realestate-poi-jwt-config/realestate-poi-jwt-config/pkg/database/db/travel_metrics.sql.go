// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: travel_metrics.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"
)

const saveTravelMetrics = `-- name: SaveTravelMetrics :execresult
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
    updated_at = CURRENT_TIMESTAMP
`

type SaveTravelMetricsParams struct {
	Distance      int32  `db:"distance" json:"distance"`
	Duration      int32  `db:"duration" json:"duration"`
	OriginID      string `db:"origin_id" json:"origin_id"`
	DestinationID int32  `db:"destination_id" json:"destination_id"`
}

func (q *Queries) SaveTravelMetrics(ctx context.Context, arg SaveTravelMetricsParams) (pgconn.CommandTag, error) {
	return q.db.Exec(ctx, saveTravelMetrics,
		arg.Distance,
		arg.Duration,
		arg.OriginID,
		arg.DestinationID,
	)
}
