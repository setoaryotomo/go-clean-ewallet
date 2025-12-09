package repositories

import (
	"database/sql"
	"sample/models"
)

// ProductRepository
type ProductRepository interface {
	FindProductById(id int) (models.Product, error)
	FindProductByIndex(code string) (models.Product, error)
	IsProductExistsByIndex(code string) (models.Product, bool)
	AddProduct(product models.Product) (int, error)
	UpdateProduct(product models.Product) (int, error)
	RemoveProduct(id int) error
	GetProductList() ([]models.Product, error)
}

// WarehouseRepository
type WarehouseRepository interface {
	FindWarehouseById(id int) (models.Warehouse, error)
}

// ProductMongoRepository
type ProductMongoRepository interface {
	AddProductMongo(product models.Product) error
}

// PedagangKiosPoinRepository
type PedagangKiosPoinRepository interface {
	FindPedagangKiosDataByDate(date string) ([]models.PedagangKiosPoin, error)
	AddPedagangKiosPoin([]models.PedagangKiosPoin) (bool, error)
}

// PedagangKiosGradingRepository
type PedagangKiosGradingRepository interface {
	FindPedagangKiosGradingWeekly(idcorporate, week int) ([]models.PedagangKiosGradingWeekly, error)
	AddPedagangKiosGradingWeekly([]models.ResponseFindPedagangKiosGradingWeekly) (bool, error)
	FindWeekPoinBonus() ([]models.WeekPoinBonus, error)
}

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
}
