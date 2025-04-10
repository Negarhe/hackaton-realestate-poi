// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: poi.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const get3NearestPoiFromEachType = `-- name: Get3NearestPoiFromEachType :many
WITH temp AS (
    SELECT
    p.id AS poi_id,
    p.name AS poi_name,
    p.type AS poi_type,
    p.address AS poi_address,
    p.latitude AS poi_latitude,
    p.longitude AS poi_longitude,
    tm.distance AS distance,
    tm.duration AS duration,
    ROW_NUMBER() OVER(PARTITION BY p.type ORDER BY tm.distance ASC) AS rank
    FROM travel_metrics tm
    JOIN poi p ON p.id = tm.destination_id
    JOIN posts pt ON pt.post_id = tm.origin_id
    WHERE 
     pt.latitude = $1 AND 
     pt.longitude = $2 
)
SELECT poi_id, poi_name, poi_type, poi_address, poi_latitude, poi_longitude, distance, duration, rank
FROM temp
WHERE rank <=3
`

type Get3NearestPoiFromEachTypeParams struct {
	Latitude  float64 `db:"latitude" json:"latitude"`
	Longitude float64 `db:"longitude" json:"longitude"`
}

type Get3NearestPoiFromEachTypeRow struct {
	PoiID        int32       `db:"poi_id" json:"poi_id"`
	PoiName      string      `db:"poi_name" json:"poi_name"`
	PoiType      PoiType     `db:"poi_type" json:"poi_type"`
	PoiAddress   pgtype.Text `db:"poi_address" json:"poi_address"`
	PoiLatitude  float64     `db:"poi_latitude" json:"poi_latitude"`
	PoiLongitude float64     `db:"poi_longitude" json:"poi_longitude"`
	Distance     int32       `db:"distance" json:"distance"`
	Duration     int32       `db:"duration" json:"duration"`
	Rank         int64       `db:"rank" json:"rank"`
}

func (q *Queries) Get3NearestPoiFromEachType(ctx context.Context, arg Get3NearestPoiFromEachTypeParams) ([]Get3NearestPoiFromEachTypeRow, error) {
	rows, err := q.db.Query(ctx, get3NearestPoiFromEachType, arg.Latitude, arg.Longitude)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Get3NearestPoiFromEachTypeRow
	for rows.Next() {
		var i Get3NearestPoiFromEachTypeRow
		if err := rows.Scan(
			&i.PoiID,
			&i.PoiName,
			&i.PoiType,
			&i.PoiAddress,
			&i.PoiLatitude,
			&i.PoiLongitude,
			&i.Distance,
			&i.Duration,
			&i.Rank,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getOrCreatePoi = `-- name: GetOrCreatePoi :one
INSERT INTO poi(name, type, latitude, longitude)
VALUES ($1,$2,$3,$4)
ON CONFLICT (name, type) DO NOTHING
RETURNING id, name, address, type, latitude, longitude, created_at
`

type GetOrCreatePoiParams struct {
	Name      string  `db:"name" json:"name"`
	Type      PoiType `db:"type" json:"type"`
	Latitude  float64 `db:"latitude" json:"latitude"`
	Longitude float64 `db:"longitude" json:"longitude"`
}

func (q *Queries) GetOrCreatePoi(ctx context.Context, arg GetOrCreatePoiParams) (Poi, error) {
	row := q.db.QueryRow(ctx, getOrCreatePoi,
		arg.Name,
		arg.Type,
		arg.Latitude,
		arg.Longitude,
	)
	var i Poi
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Address,
		&i.Type,
		&i.Latitude,
		&i.Longitude,
		&i.CreatedAt,
	)
	return i, err
}

const upsertPOI = `-- name: UpsertPOI :one
INSERT INTO poi(name, type, latitude, longitude,address)
VALUES ($1,$2,$3,$4,$5)
ON CONFLICT (type, latitude, longitude)
DO UPDATE SET name = EXCLUDED.name
RETURNING id
`

type UpsertPOIParams struct {
	Name      string      `db:"name" json:"name"`
	Type      PoiType     `db:"type" json:"type"`
	Latitude  float64     `db:"latitude" json:"latitude"`
	Longitude float64     `db:"longitude" json:"longitude"`
	Address   pgtype.Text `db:"address" json:"address"`
}

func (q *Queries) UpsertPOI(ctx context.Context, arg UpsertPOIParams) (int32, error) {
	row := q.db.QueryRow(ctx, upsertPOI,
		arg.Name,
		arg.Type,
		arg.Latitude,
		arg.Longitude,
		arg.Address,
	)
	var id int32
	err := row.Scan(&id)
	return id, err
}
