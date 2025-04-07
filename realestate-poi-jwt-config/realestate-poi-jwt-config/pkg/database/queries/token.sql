-- name: GetAccessTokenByUserIdPostId :one
SELECT * FROM tokens t
JOIN posts p ON p.post_id=t.post_id
JOIN users u ON u.id=t.user_id
WHERE  u.id = $1 AND 
        p.post_id = $2
LIMIT 1;

-- name: InsertToken :execresult
INSERT INTO tokens (post_id, user_id, access_token, refresh_token, expires_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (post_id, user_id) DO UPDATE
SET
    access_token = EXCLUDED.access_token,
    refresh_token = EXCLUDED.refresh_token,
    expires_at= EXCLUDED.expires_at;
-- WHERE now() > tokens.expires_at;