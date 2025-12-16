package repositories

import (
	"database/sql"
	"sample/models"
)

// AccountRepository
type AccountRepository interface {
	FindAccountById(id int) (models.Account, error)
	FindAccountByNumber(accountNumber string) (models.Account, error)
	IsAccountExistsByNumber(accountNumber string) (models.Account, bool)
	AddAccount(account models.Account) (int, error)
	UpdateAccount(account models.Account) (int, error)
	UpdatePIN(accountNumber string, newPIN string) error
	UpdatePINWithTx(tx *sql.Tx, accountNumber string, newPIN string) error
	ChangePINWithTx(tx *sql.Tx, accountNumber, oldPIN, newPIN string, currentHashedPIN string) (int, error)
	IncrementFailedPINAttempts(accountNumber string) (int, error)
	ResetFailedPINAttempts(accountNumber string) error
	IncrementDecrementLastBalance(accountID int, amount float64, debitCreditOperator string, updatedAt string, tx *sql.Tx) (lastBalance float64, err error)
	RemoveAccount(id int) error
	GetAccountList() ([]models.Account, error)
	VerifyPIN(accountNumber string, pin string) (bool, error)
}

// TransactionRepository
type TransactionRepository interface {
	AddTransaction(transaction models.Transaction) (int, error)
	FindTransactionById(id int) (models.Transaction, error)
	GetTransactionHistory(accountNumber string, startDate, endDate string, limit, page int) ([]models.Transaction, int, error)
	DataCountAndSumTransactionListByIndex(countOnly bool, filter models.RequestTransactionHistoryList) (models.ResultDataTableTransactionCountAndSummaries, error)
	DataGetTransactionListByIndex(filter models.RequestTransactionHistoryList) ([]models.Transaction, error)
}
