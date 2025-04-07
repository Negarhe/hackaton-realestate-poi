CREATE TABLE post_purchase(
    user_id VARCHAR(255) NOT NULL,
    post_token VARCHAR(255) NOT NULL,
    purchased_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    -- expires_at TIMESTAMP NOT NULL,
    PRIMARY KEY (user_id, post_token),
    FOREIGN KEY(user_id) REFERENCES users(id),
    FOREIGN KEY (post_token) REFERENCES posts(post_id)
);