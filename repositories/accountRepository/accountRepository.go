package accountRepository

import (
	"database/sql"
	"errors"
	"sample/helpers"
	"sample/models"
	"sample/repositories"
	"time"
)

var defineColumn = `id, account_number, balance, pin, account_name, account_status, failed_pin_attempts, created_at, updated_at`

type accountRepository struct {
	RepoDB repositories.Repository
}

// NewAccountRepository
func NewAccountRepository(repoDB repositories.Repository) accountRepository {
	return accountRepository{
		RepoDB: repoDB,
	}
}

// FindAccountById mencari berdasarkan id
func (ctx accountRepository) FindAccountById(id int) (models.Account, error) {
	var account models.Account

	var query = `SELECT ` + defineColumn + ` FROM account WHERE id = $1 AND deleted_at IS NULL`

	err := ctx.RepoDB.DB.QueryRow(query, id).Scan(
		&account.ID,
		&account.AccountNumber,
		&account.Balance,
		&account.PIN,
		&account.AccountName,
		&account.AccountStatus,
		&account.FailedPINAttempts,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return account, errors.New("Account not found")
		}
		return account, err
	}

	return account, nil
}

// FindAccountByNumber mencari berdasarkan nomor akun
func (ctx accountRepository) FindAccountByNumber(accountNumber string) (models.Account, error) {
	var account models.Account

	var query = `SELECT ` + defineColumn + ` FROM account WHERE account_number = $1 AND deleted_at IS NULL`

	err := ctx.RepoDB.DB.QueryRow(query, accountNumber).Scan(
		&account.ID,
		&account.AccountNumber,
		&account.Balance,
		&account.PIN,
		&account.AccountName,
		&account.AccountStatus,
		&account.FailedPINAttempts,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return account, errors.New("Account not found")
		}
		return account, err
	}

	return account, nil
}

// IsAccountExistsByNumber cek apakah akun ada
func (ctx accountRepository) IsAccountExistsByNumber(accountNumber string) (models.Account, bool) {
	var account models.Account

	var query = `SELECT ` + defineColumn + ` FROM account WHERE account_number = $1 AND deleted_at IS NULL`

	err := ctx.RepoDB.DB.QueryRow(query, accountNumber).Scan(
		&account.ID,
		&account.AccountNumber,
		&account.Balance,
		&account.PIN,
		&account.AccountName,
		&account.AccountStatus,
		&account.FailedPINAttempts,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		return account, false
	}

	return account, true
}

// AddAccount membuat akun baru
func (ctx accountRepository) AddAccount(account models.Account) (int, error) {
	var ID int

	query := `INSERT INTO account (
				account_number, balance, pin, account_name, account_status, failed_pin_attempts, created_at, updated_at
		) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8
		) RETURNING id`

	now := time.Now()
	err := ctx.RepoDB.DB.QueryRow(
		query,
		account.AccountNumber,
		account.Balance,
		account.PIN,
		account.AccountName,
		"ACTIVE", // Default status
		0,        // Default failed attempts
		now,
		now,
	).Scan(&ID)

	if err != nil {
		return 0, err
	}

	return ID, nil
}

// UpdateAccount update data akun
func (ctx accountRepository) UpdateAccount(account models.Account) (int, error) {
	var strQuery = `UPDATE account 
					SET account_name = $2, 
					    updated_at = $3
					WHERE id = $1 AND deleted_at IS NULL
					RETURNING id`

	var ID int
	err := ctx.RepoDB.DB.QueryRow(strQuery, account.ID, account.AccountName, time.Now()).Scan(&ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("Account not found")
		}
		return 0, err
	}

	return ID, nil
}

// UpdatePIN update PIN akun
func (ctx accountRepository) UpdatePIN(accountNumber string, newPIN string) error {
	query := `UPDATE account 
			  SET pin = $1, 
			      failed_pin_attempts = 0,
			      account_status = 'ACTIVE',
			      updated_at = $2
			  WHERE account_number = $3 AND deleted_at IS NULL`

	result, err := ctx.RepoDB.DB.Exec(query, newPIN, time.Now(), accountNumber)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("Account not found")
	}

	return nil
}

// IncrementFailedPINAttempts increment failed PIN attempts dan block jika >= 3
func (ctx accountRepository) IncrementFailedPINAttempts(accountNumber string) (int, error) {
	var failedAttempts int

	query := `UPDATE account 
			  SET failed_pin_attempts = failed_pin_attempts + 1,
			      account_status = CASE 
				      WHEN failed_pin_attempts + 1 >= 3 THEN 'BLOCKED_PIN'
				      ELSE account_status 
			      END,
			      updated_at = $1
			  WHERE account_number = $2 AND deleted_at IS NULL
			  RETURNING failed_pin_attempts`

	err := ctx.RepoDB.DB.QueryRow(query, time.Now(), accountNumber).Scan(&failedAttempts)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("Account not found")
		}
		return 0, err
	}

	return failedAttempts, nil
}

// ResetFailedPINAttempts reset failed PIN attempts
func (ctx accountRepository) ResetFailedPINAttempts(accountNumber string) error {
	query := `UPDATE account 
			  SET failed_pin_attempts = 0,
			      account_status = 'ACTIVE',
			      updated_at = $1
			  WHERE account_number = $2 AND deleted_at IS NULL`

	result, err := ctx.RepoDB.DB.Exec(query, time.Now(), accountNumber)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("Account not found")
	}

	return nil
}

// IncrementDecrementLastBalance update saldo akun dengan operator (+/-) dan return last balance
func (ctx accountRepository) IncrementDecrementLastBalance(accountID int, amount float64, debitCreditOperator string, updatedAt string, tx *sql.Tx) (lastBalance float64, err error) {
	var args []interface{}

	if debitCreditOperator != "+" && debitCreditOperator != "-" {
		return 0, errors.New("Invalid operator. Must be '+' or '-'")
	}

	query := `
		UPDATE account SET balance = ROUND((balance ` + debitCreditOperator + ` $1)::numeric, 2),
			updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
		RETURNING balance
	`

	query = helpers.ReplaceSQL(query, "?")
	args = append(args, amount, updatedAt, accountID)

	if tx != nil {
		err = tx.QueryRow(query, args...).Scan(&lastBalance)
	} else {
		err = ctx.RepoDB.DB.QueryRow(query, args...).Scan(&lastBalance)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("Account not found")
		}
		return 0, err
	}

	return lastBalance, nil
}

// RemoveAccount soft delete akun
func (ctx accountRepository) RemoveAccount(id int) error {
	result, err := ctx.RepoDB.DB.Exec(
		"UPDATE account SET deleted_at = $1 WHERE id = $2",
		time.Now(),
		id,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("Account not found")
	}

	return nil
}

// GetAccountList mendapatkan list semua akun
func (ctx accountRepository) GetAccountList() ([]models.Account, error) {
	var query = `SELECT ` + defineColumn + ` FROM account WHERE deleted_at IS NULL ORDER BY created_at DESC`

	rows, err := ctx.RepoDB.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return accountDto(rows)
}

// GetAccountBalance mendapatkan saldo akun
func (ctx accountRepository) GetAccountBalance(accountNumber string) (float64, error) {
	var balance float64

	query := `SELECT balance FROM account WHERE account_number = $1 AND deleted_at IS NULL`

	err := ctx.RepoDB.DB.QueryRow(query, accountNumber).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("Account not found")
		}
		return 0, err
	}

	return balance, nil
}

// VerifyPIN verifikasi PIN akun
func (ctx accountRepository) VerifyPIN(accountNumber string, pin string) (bool, error) {
	var storedPIN string

	query := `SELECT pin FROM account WHERE account_number = $1 AND deleted_at IS NULL`

	err := ctx.RepoDB.DB.QueryRow(query, accountNumber).Scan(&storedPIN)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, errors.New("Account not found")
		}
		return false, err
	}

	return storedPIN == pin, nil
}

// accountDto helper untuk mapping rows ke struct
func accountDto(rows *sql.Rows) ([]models.Account, error) {
	var result []models.Account

	for rows.Next() {
		var val models.Account
		err := rows.Scan(
			&val.ID,
			&val.AccountNumber,
			&val.Balance,
			&val.PIN,
			&val.AccountName,
			&val.AccountStatus,
			&val.FailedPINAttempts,
			&val.CreatedAt,
			&val.UpdatedAt,
		)
		if err != nil {
			return result, err
		}
		result = append(result, val)
	}

	return result, nil
}
