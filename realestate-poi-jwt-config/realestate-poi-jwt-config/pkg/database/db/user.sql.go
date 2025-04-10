// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: user.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"
)

const insertUser = `-- name: InsertUser :execresult
INSERT INTO users (id)
VALUES ($1)
ON CONFLICT (id) DO NOTHING
`

func (q *Queries) InsertUser(ctx context.Context, id string) (pgconn.CommandTag, error) {
	return q.db.Exec(ctx, insertUser, id)
}
