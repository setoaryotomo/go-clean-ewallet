package transactionRepository

import (
	"database/sql"
	"errors"
	"sample/models"
	"sample/repositories"
	"strconv"
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
		nullString(transaction.SourceNumber),
		nullString(transaction.BeneficiaryNumber),
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

// GetTransactionHistory mendapatkan riwayat transaksi dengan pagination
func (ctx transactionRepository) GetTransactionHistory(accountNumber string, startDate, endDate string, limit, page int) ([]models.Transaction, int, error) {
	var totalRecords int

	// Set default limit dan page
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	// Base query
	baseQuery := `FROM transaction WHERE account_number = $1 AND deleted_at IS NULL`
	params := []interface{}{accountNumber}
	paramCount := 1

	// Filter berdasarkan tanggal jika ada
	if startDate != "" && endDate != "" {
		paramCount++
		baseQuery += ` AND DATE(transaction_time) BETWEEN $` + strconv.Itoa(paramCount) + ` AND $` + strconv.Itoa(paramCount+1)
		params = append(params, startDate, endDate)
		paramCount++
	}

	// Count total records
	countQuery := `SELECT COUNT(*) ` + baseQuery
	err := ctx.RepoDB.DB.QueryRow(countQuery, params...).Scan(&totalRecords)
	if err != nil {
		return nil, 0, err
	}

	// Get data dengan pagination
	dataQuery := `SELECT ` + defineColumn + ` ` + baseQuery + ` ORDER BY transaction_time DESC LIMIT $` +
		strconv.Itoa(paramCount+1) + ` OFFSET $` + strconv.Itoa(paramCount+2)
	params = append(params, limit, offset)

	rows, err := ctx.RepoDB.DB.Query(dataQuery, params...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	transactions, err := transactionDto(rows)
	if err != nil {
		return nil, 0, err
	}

	return transactions, totalRecords, nil
}

// transactionDto helper untuk mapping rows ke struct
func transactionDto(rows *sql.Rows) ([]models.Transaction, error) {
	var result []models.Transaction

	for rows.Next() {
		var val models.Transaction
		var sourceNumber, beneficiaryNumber sql.NullString

		err := rows.Scan(
			&val.ID,
			&val.AccountID,
			&val.AccountNumber,
			&val.AccountName,
			&sourceNumber,
			&beneficiaryNumber,
			&val.TransactionType,
			&val.Amount,
			&val.TransactionTime,
			&val.CreatedAt,
		)
		if err != nil {
			return result, err
		}

		val.SourceNumber = sourceNumber.String
		val.BeneficiaryNumber = beneficiaryNumber.String

		result = append(result, val)
	}

	return result, nil
}

// Helper functions
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
