package transactionRepository

import (
	"database/sql"
	"errors"
	"sample/helpers"
	"sample/models"
	"sample/repositories"
	"time"
)

var defineColumn = `id, account_id, account_number, account_name, source_number, 
					beneficiary_number, transaction_type, amount, transaction_time, created_at`

type transactionRepository struct {
	RepoDB repositories.Repository
}

// NewTransactionRepository
func NewTransactionRepository(repoDB repositories.Repository) transactionRepository {
	return transactionRepository{
		RepoDB: repoDB,
	}
}

// AddTransaction mencatat transaksi baru
func (ctx transactionRepository) AddTransaction(transaction models.Transaction) (int, error) {
	var ID int

	query := `INSERT INTO transaction (
				account_id, account_number, account_name, 
				source_number, beneficiary_number,
				transaction_type, amount, transaction_time, created_at
		) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING id`

	now := time.Now()
	err := ctx.RepoDB.DB.QueryRow(
		query,
		transaction.AccountID,
		transaction.AccountNumber,
		transaction.AccountName,
		helpers.NullString(transaction.SourceNumber),
		helpers.NullString(transaction.BeneficiaryNumber),
		transaction.TransactionType,
		transaction.Amount,
		transaction.TransactionTime,
		now,
	).Scan(&ID)

	if err != nil {
		return 0, err
	}

	return ID, nil
}

// FindTransactionById mencari transaksi berdasarkan ID
func (ctx transactionRepository) FindTransactionById(id int) (models.Transaction, error) {
	var transaction models.Transaction

	var query = `SELECT ` + defineColumn + ` FROM transaction WHERE id = $1 AND deleted_at IS NULL`

	var sourceNumber, beneficiaryNumber sql.NullString

	err := ctx.RepoDB.DB.QueryRow(query, id).Scan(
		&transaction.ID,
		&transaction.AccountID,
		&transaction.AccountNumber,
		&transaction.AccountName,
		&sourceNumber,
		&beneficiaryNumber,
		&transaction.TransactionType,
		&transaction.Amount,
		&transaction.TransactionTime,
		&transaction.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return transaction, errors.New("Transaction not found")
		}
		return transaction, err
	}

	transaction.SourceNumber = sourceNumber.String
	transaction.BeneficiaryNumber = beneficiaryNumber.String

	return transaction, nil
}

// GetTransactionHistory mendapatkan riwayat transaksi
func (ctx transactionRepository) GetTransactionHistory(accountNumber string, startDate, endDate string, limit, page int) ([]models.Transaction, int, error) {
	var totalRecords int

	// Base query dan parameter
	var baseQuery string
	var params []interface{}
	paramCount := 1

	// Filter berdasarkan account number
	if accountNumber != "" {
		baseQuery = `FROM transaction WHERE account_number = $1 AND deleted_at IS NULL`
		params = append(params, accountNumber)
	} else {
		baseQuery = `FROM transaction WHERE deleted_at IS NULL`
		paramCount = 0
	}

	// Filter berdasarkan tanggal
	if startDate != "" && endDate != "" {
		if baseQuery == `FROM transaction WHERE deleted_at IS NULL` {
			baseQuery += ` AND DATE(transaction_time) BETWEEN $1 AND $2`
			params = append(params, startDate, endDate)
			paramCount = 2
		} else {
			baseQuery += ` AND DATE(transaction_time) BETWEEN $2 AND $3`
			params = append(params, startDate, endDate)
			paramCount = 3
		}
	} else if startDate != "" || endDate != "" {
		if startDate != "" {
			if baseQuery == `FROM transaction WHERE deleted_at IS NULL` {
				baseQuery += ` AND DATE(transaction_time) >= $1`
				params = append(params, startDate)
				paramCount = 1
			} else {
				baseQuery += ` AND DATE(transaction_time) >= $2`
				params = append(params, startDate)
				paramCount = 2
			}
		}
		if endDate != "" {
			if baseQuery == `FROM transaction WHERE deleted_at IS NULL` {
				if startDate != "" {
					baseQuery += ` AND DATE(transaction_time) <= $2`
				} else {
					baseQuery += ` AND DATE(transaction_time) <= $1`
				}
				params = append(params, endDate)
				if startDate != "" {
					paramCount = 2
				} else {
					paramCount = 1
				}
			} else {
				if startDate != "" {
					baseQuery += ` AND DATE(transaction_time) <= $3`
					paramCount = 3
				} else {
					baseQuery += ` AND DATE(transaction_time) <= $2`
					paramCount = 2
				}
				params = append(params, endDate)
			}
		}
	}

	// Count total records
	countQuery := `SELECT COUNT(*) ` + baseQuery
	err := ctx.RepoDB.DB.QueryRow(countQuery, params...).Scan(&totalRecords)
	if err != nil {
		return nil, 0, err
	}

	// Get data
	dataQuery := `SELECT ` + defineColumn + ` ` + baseQuery + ` ORDER BY transaction_time DESC`

	// Jika limit > 0, gunakan pagination
	if limit > 0 {
		if page <= 0 {
			page = 1
		}
		offset := (page - 1) * limit

		// Tambahkan parameter untuk limit dan offset
		if paramCount == 0 {
			dataQuery += ` LIMIT $1 OFFSET $2`
			params = append(params, limit, offset)
		} else if paramCount == 1 {
			dataQuery += ` LIMIT $2 OFFSET $3`
			params = append(params, limit, offset)
		} else if paramCount == 2 {
			dataQuery += ` LIMIT $3 OFFSET $4`
			params = append(params, limit, offset)
		} else if paramCount == 3 {
			dataQuery += ` LIMIT $4 OFFSET $5`
			params = append(params, limit, offset)
		}
	}
	// Jika limit = 0, tidak ada LIMIT dan OFFSET (ambil semua data)

	rows, err := ctx.RepoDB.DB.Query(dataQuery, params...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	transactions, err := helpers.TransactionDto(rows)
	if err != nil {
		return nil, 0, err
	}

	return transactions, totalRecords, nil
}
