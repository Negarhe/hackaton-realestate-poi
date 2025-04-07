CREATE TYPE poi_type AS ENUM ('subway', 'hospital', 'super_market', 'bus_station', 'fruit_market');

-- Create the poi table
CREATE TABLE poi (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address VARCHAR(255) ,
    type poi_type NOT NULL, -- 'subway', 'hospital', 'mall' 
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP  DEFAULT CURRENT_TIMESTAMP,
    UNIQUE  (type, latitude, longitude) -- before i used to have name, type but it does not allow 2 hospital with same name in different cities, but this constrain prevents duplicate
);

