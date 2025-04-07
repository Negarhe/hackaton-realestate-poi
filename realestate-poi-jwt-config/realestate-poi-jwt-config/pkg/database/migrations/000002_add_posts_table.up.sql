-- Create the posts table
CREATE TABLE posts (
    post_id VARCHAR(255) PRIMARY KEY,
    owner_id VARCHAR REFERENCES users(id),
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    title VARCHAR(255),
    -- coordinates_set BOOLEAN DEFAULT FALSE,
    -- access_token TEXT  NOT NULL,
    -- refresh_token TEXT NOT NULL,
    -- expires_in TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    CONSTRAINT no_empty_id CHECK (post_id <> '')
    -- FOREIGN KEY(user_id) REFERENCES users(id)
    -- CONSTRAINT coordinates_complete CHECK ((latitude IS NOT NULL AND longitude IS NOT NULL AND coordinates_set=TRUE) OR (latitude IS NULL AND longitude IS NULL AND coordinates_set=FALSE))
);


