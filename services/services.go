package services

import (
	"database/sql"
	"sample/repositories"
)

type UsecaseService struct {
	WarehouseRepo           repositories.WarehouseRepository
	ProductRepo             repositories.ProductRepository
	RepoDB                  *sql.DB
	ProductMongoRepo        repositories.ProductMongoRepository
	PedagangKiosPoinRepo    repositories.PedagangKiosPoinRepository
	PedagangKiosGradingRepo repositories.PedagangKiosGradingRepository
	AccountRepo             repositories.AccountRepository
	TransactionRepo         repositories.TransactionRepository
}

func NewUsecaseService(repoDB *sql.DB,
	WarehouseRepo repositories.WarehouseRepository,
	ProductRepo repositories.ProductRepository,
	ProductMongoRepo repositories.ProductMongoRepository,
	PedagangKiosPoinRepo repositories.PedagangKiosPoinRepository,
	PedagangKiosGradingRepo repositories.PedagangKiosGradingRepository,
	AccountRepo repositories.AccountRepository,
	TransactionRepo repositories.TransactionRepository,
) UsecaseService {
	return UsecaseService{
		RepoDB:                  repoDB,
		WarehouseRepo:           WarehouseRepo,
		ProductRepo:             ProductRepo,
		AccountRepo:             AccountRepo,
		TransactionRepo:         TransactionRepo,
		ProductMongoRepo:        ProductMongoRepo,
		PedagangKiosPoinRepo:    PedagangKiosPoinRepo,
		PedagangKiosGradingRepo: PedagangKiosGradingRepo,
	}
}
