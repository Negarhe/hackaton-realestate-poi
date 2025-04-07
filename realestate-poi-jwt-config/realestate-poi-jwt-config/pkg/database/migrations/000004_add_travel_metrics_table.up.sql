Create TABLE travel_metrics(
    id SERIAL PRIMARY KEY,
    distance INT NOT NULL,
    duration INT NOT NULL,
    origin_id VARCHAR(255) NOT NULL,
    destination_id INTEGER NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(origin_id) REFERENCES posts(post_id),
    FOREIGN KEY(destination_id) REFERENCES poi(id),
    UNIQUE(origin_id, destination_id)
);