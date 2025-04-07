-- name: InsertPostPurchase :execresult
INSERT INTO post_purchase (user_id,post_token)
VALUES ($1, $2)
ON CONFLICT (user_id, post_token)
DO UPDATE SET
    purchased_at = CURRENT_TIMESTAMP;

-- name: CheckUserPurchase :one
SELECT 
    EXISTS (
        SELECT 1 
        FROM post_purchase 
        WHERE user_id = $1 AND post_token = $2
    ) AS has_purchased;
