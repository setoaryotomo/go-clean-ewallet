package repositories

import "sample/models"

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

// AccountRepository
type AccountRepository interface {
	FindAccountById(id int) (models.Account, error)
	FindAccountByNumber(accountNumber string) (models.Account, error)
	IsAccountExistsByNumber(accountNumber string) (models.Account, bool)
	AddAccount(account models.Account) (int, error)
	UpdateAccount(account models.Account) (int, error)
	UpdateBalance(accountNumber string, amount float64) error
	UpdatePIN(accountNumber string, newPIN string) error
	RemoveAccount(id int) error
	GetAccountList() ([]models.Account, error)
	GetAccountBalance(accountNumber string) (float64, error)
	VerifyPIN(accountNumber string, pin string) (bool, error)
}

// WarehouseRepository
type WarehouseRepository interface {
	FindWarehouseById(id int) (models.Warehouse, error)
}

// ProductMongoRepository
type ProductMongoRepository interface {
	AddProductMongo(product models.Product) error
}

type PedagangKiosPoinRepository interface {
	FindPedagangKiosDataByDate(date string) ([]models.PedagangKiosPoin, error)
	AddPedagangKiosPoin([]models.PedagangKiosPoin) (bool, error)
}

type PedagangKiosGradingRepository interface {
	FindPedagangKiosGradingWeekly(idcorporate, week int) ([]models.PedagangKiosGradingWeekly, error)
	AddPedagangKiosGradingWeekly([]models.ResponseFindPedagangKiosGradingWeekly) (bool, error)
	FindWeekPoinBonus() ([]models.WeekPoinBonus, error)
}
