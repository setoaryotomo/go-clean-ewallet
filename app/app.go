package app

import (
	"database/sql"
	"sample/repositories"
	"sample/repositories/accountRepository"
	"sample/repositories/transactionRepository"
	"sample/services"
)

func SetupApp(DB *sql.DB, repo repositories.Repository) services.UsecaseService {

	// Repository
	accountRepo := accountRepository.NewAccountRepository(repo)
	transactionRepo := transactionRepository.NewTransactionRepository(repo)

	// Services
	usecaseSvc := services.NewUsecaseService(DB, accountRepo, transactionRepo)

	return usecaseSvc
}
