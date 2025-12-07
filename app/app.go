package app

import (
	"database/sql"
	"sample/repositories"
	"sample/repositories/accountRepository"
	"sample/repositories/pedagangKiosGradingRepository"
	"sample/repositories/pedagangKiosPoinRepository"
	"sample/repositories/productRepository"
	"sample/repositories/transactionRepository"
	"sample/repositories/warehouseRepository"
	"sample/services"
)

func SetupApp(DB *sql.DB, repo repositories.Repository) services.UsecaseService {

	// Repository
	productRepo := productRepository.NewProductRepository(repo)
	productMongoRepo := productRepository.NewProductMongoRepository(repo)
	warehouseRepo := warehouseRepository.NewWarehouseRepository(repo)
	pedagangKiosPoinRepo := pedagangKiosPoinRepository.NewPedagangKiosPoinRepository(repo)
	pedagangKiosGradingRepo := pedagangKiosGradingRepository.NewPedagangKiosGradingRepository(repo)
	accountRepo := accountRepository.NewAccountRepository(repo)
	transactionRepo := transactionRepository.NewTransactionRepository(repo)

	// Services
	usecaseSvc := services.NewUsecaseService(DB, warehouseRepo, productRepo, productMongoRepo, pedagangKiosPoinRepo, pedagangKiosGradingRepo, accountRepo, transactionRepo)

	return usecaseSvc
}
