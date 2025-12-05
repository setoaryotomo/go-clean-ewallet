package repositories

import (
	"context"
	"database/sql"

	"go.mongodb.org/mongo-driver/mongo"
)

type Repository struct {
	DB      *sql.DB
	MongoDB *mongo.Database
	Context context.Context
}

func NewRepository(conn *sql.DB, MongoDB *mongo.Database, ctx context.Context) Repository {
	return Repository{
		DB:      conn,
		MongoDB: MongoDB,
		Context: ctx,
	}
}
