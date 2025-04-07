-- name: GetOrCreatePoi :one
INSERT INTO poi(name, type, latitude, longitude)
VALUES ($1,$2,$3,$4)
ON CONFLICT (name, type) DO NOTHING
RETURNING *;


-- name: Get3NearestPoiFromEachType :many
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
SELECT *
FROM temp
WHERE rank <=3;


-- name: UpsertPOI :one
INSERT INTO poi(name, type, latitude, longitude,address)
VALUES ($1,$2,$3,$4,$5)
ON CONFLICT (type, latitude, longitude)
DO UPDATE SET name = EXCLUDED.name
RETURNING id;

-- name: UpsertSubwayStation :one
-- INSERT INTO poi (name, type, latitude, longitude)
-- VALUES ($1, 'subway', $2, $3)
-- ON CONFLICT (name, type) 
-- DO UPDATE SET 
--     latitude = EXCLUDED.latitude,
--     longitude = EXCLUDED.longitude
-- RETURNING id;

-- name: GetToSubwayInfo :one
-- Get the nearest subway station to a given lat,lng (There is an issue, many posts can have the same lat,lng)
-- SELECT 
--     p.name AS station_name,
--     tm.distance,
--     tm.duration,
--     p.latitude AS station_latitude,
--     p.longitude AS station_longitude
-- FROM travel_metrics tm
-- JOIN poi p ON p.id = tm.destination_id
-- JOIN posts pt ON pt.post_id = tm.origin_id
-- WHERE 
--     pt.latitude = $1 AND 
--     pt.longitude = $2 AND
--     p.type = 'subway'
-- LIMIT 1;


-- name: GetNearestPoi :many
-- SELECT
--     p.id,
--     p.name,
--     p.type,
--     p.latitude,
--     p.longitude,
--     tm.distance,
--     tm.duration
-- FROM travel_metrics tm
-- JOIN poi p ON p.id = tm.destination_id
-- WHERE ST.DISTACE(
--     ST_SRID(ST_MakePoint($2, $1),4326),
--     ST_SRID(ST_MakePoint(p.longitude, p.latitude),4326)
-- ) < $3
-- ORDER BY p.type, tm.distance;



