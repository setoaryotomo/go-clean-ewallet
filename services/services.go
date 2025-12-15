package services

import (
	"database/sql"
	"sample/repositories"
)

type UsecaseService struct {
	RepoDB          *sql.DB
	AccountRepo     repositories.AccountRepository
	TransactionRepo repositories.TransactionRepository
}

func NewUsecaseService(repoDB *sql.DB,
	AccountRepo repositories.AccountRepository,
	TransactionRepo repositories.TransactionRepository,
) UsecaseService {
	return UsecaseService{
		RepoDB:          repoDB,
		AccountRepo:     AccountRepo,
		TransactionRepo: TransactionRepo,
	}
}
