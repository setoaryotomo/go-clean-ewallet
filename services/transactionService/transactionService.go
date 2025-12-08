package transactionService

import (
	"database/sql"
	"net/http"
	"sample/constans"
	"sample/helpers"
	"sample/models"
	"sample/services"
	"sample/utils"
	"time"

	"github.com/labstack/echo"
	"golang.org/x/crypto/bcrypt"
)

type transactionService struct {
	Service services.UsecaseService
}

func NewTransactionService(service services.UsecaseService) transactionService {
	return transactionService{
		Service: service,
	}
}

// Deposit menambah saldo
func (svc transactionService) Deposit(ctx echo.Context) error {
	var result models.Response

	request := new(models.RequestDeposit)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Cek akun exists
	account, err := svc.Service.AccountRepo.FindAccountByNumber(request.AccountNumber)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	balanceBefore := account.Balance
	transactionTime := time.Now()
	updatedAt := transactionTime.Format("2006-01-02 15:04:05")

	var balanceAfter float64

	// DB Transaction menggunakan utils.DBTransaction
	err = utils.DBTransaction(svc.Service.RepoDB, func(tx *sql.Tx) error {
		// Update balance menggunakan IncrementDecrementLastBalance dengan operator "+"
		lastBalance, err := svc.Service.AccountRepo.IncrementDecrementLastBalance(
			account.ID,
			request.Amount,
			"+", // Credit operator
			updatedAt,
			tx,
		)
		if err != nil {
			return err
		}
		balanceAfter = lastBalance

		// Record transaction (Credit)
		transaction := models.Transaction{
			AccountID:         account.ID,
			AccountNumber:     account.AccountNumber,
			AccountName:       account.AccountName,
			TransactionType:   "C", // Credit (+)
			Amount:            request.Amount,
			TransactionTime:   transactionTime,
			SourceNumber:      account.AccountNumber,
			BeneficiaryNumber: account.AccountNumber,
		}

		_, err = svc.Service.TransactionRepo.AddTransaction(transaction)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, "Transaction failed: "+err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	response := map[string]interface{}{
		"account_number":   account.AccountNumber,
		"account_name":     account.AccountName,
		"balance_before":   balanceBefore,
		"amount":           request.Amount,
		"balance_after":    balanceAfter,
		"transaction_date": transactionTime,
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "Deposit successful", response)
	return ctx.JSON(http.StatusOK, result)
}

// Withdraw menarik saldo
func (svc transactionService) Withdraw(ctx echo.Context) error {
	var result models.Response

	request := new(models.RequestWithdraw)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Verifikasi PIN
	account, err := svc.Service.AccountRepo.FindAccountByNumber(request.AccountNumber)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	if !checkPINHash(request.PIN, account.PIN) {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Invalid PIN", nil)
		return ctx.JSON(http.StatusUnauthorized, result)
	}

	// Cek saldo cukup
	if account.Balance < request.Amount {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Insufficient balance", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	balanceBefore := account.Balance
	transactionTime := time.Now()
	updatedAt := transactionTime.Format("2006-01-02 15:04:05")

	var balanceAfter float64

	// DB Transaction menggunakan utils.DBTransaction
	err = utils.DBTransaction(svc.Service.RepoDB, func(tx *sql.Tx) error {
		// Update balance menggunakan IncrementDecrementLastBalance dengan operator "-"
		lastBalance, err := svc.Service.AccountRepo.IncrementDecrementLastBalance(
			account.ID,
			request.Amount,
			"-", // Debit operator
			updatedAt,
			tx,
		)
		if err != nil {
			return err
		}
		balanceAfter = lastBalance

		// Check if balance after is negative
		if balanceAfter < 0 {
			return &utils.TransactionError{
				Code:    constans.ACCOUNT_BALANCE_BELOW_MINIMUM_CODE,
				Message: "Account balance below minimum",
			}
		}

		// Record transaction (Debit)
		transaction := models.Transaction{
			AccountID:         account.ID,
			AccountNumber:     account.AccountNumber,
			AccountName:       account.AccountName,
			TransactionType:   "D", // Debit (-)
			Amount:            request.Amount,
			TransactionTime:   transactionTime,
			SourceNumber:      account.AccountNumber,
			BeneficiaryNumber: account.AccountNumber,
		}

		_, err = svc.Service.TransactionRepo.AddTransaction(transaction)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		// Handle custom transaction errors
		if txErr, ok := err.(*utils.TransactionError); ok {
			if txErr.Code == constans.ACCOUNT_BALANCE_BELOW_MINIMUM_CODE {
				result = helpers.ResponseJSON(false, constans.ACCOUNT_BALANCE_BELOW_MINIMUM_CODE, txErr.Message, nil)
				return ctx.JSON(http.StatusBadRequest, result)
			}
		}

		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, "Transaction failed: "+err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	response := map[string]interface{}{
		"account_number":   account.AccountNumber,
		"account_name":     account.AccountName,
		"balance_before":   balanceBefore,
		"amount":           request.Amount,
		"balance_after":    balanceAfter,
		"transaction_date": transactionTime,
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "Withdraw successful", response)
	return ctx.JSON(http.StatusOK, result)
}

// Transfer antar akun
func (svc transactionService) Transfer(ctx echo.Context) error {
	var result models.Response

	request := new(models.RequestTransfer)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Validasi tidak bisa transfer ke akun sendiri
	if request.FromAccountNumber == request.ToAccountNumber {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Cannot transfer to same account", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Get akun pengirim
	fromAccount, err := svc.Service.AccountRepo.FindAccountByNumber(request.FromAccountNumber)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Source account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	// Verifikasi PIN
	if !checkPINHash(request.PIN, fromAccount.PIN) {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Invalid PIN", nil)
		return ctx.JSON(http.StatusUnauthorized, result)
	}

	// Cek saldo cukup
	if fromAccount.Balance < request.Amount {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Insufficient balance", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Get akun penerima
	toAccount, err := svc.Service.AccountRepo.FindAccountByNumber(request.ToAccountNumber)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Beneficiary account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	fromBalanceBefore := fromAccount.Balance
	toBalanceBefore := toAccount.Balance
	transactionTime := time.Now()
	updatedAt := transactionTime.Format("2006-01-02 15:04:05")

	var fromBalanceAfter, toBalanceAfter float64

	// DB Transaction menggunakan utils.DBTransaction
	err = utils.DBTransaction(svc.Service.RepoDB, func(tx *sql.Tx) error {
		// Kurangi saldo pengirim dengan operator "-"
		lastBalance, err := svc.Service.AccountRepo.IncrementDecrementLastBalance(
			fromAccount.ID,
			request.Amount,
			"-", // Debit operator
			updatedAt,
			tx,
		)
		if err != nil {
			return err
		}
		fromBalanceAfter = lastBalance

		// Check if sender's balance after is negative
		if fromBalanceAfter < 0 {
			return &utils.TransactionError{
				Code:    constans.ACCOUNT_BALANCE_BELOW_MINIMUM_CODE,
				Message: "Sender balance would be negative after transfer",
			}
		}

		// Record transaksi Debit untuk pengirim
		debitTransaction := models.Transaction{
			AccountID:         fromAccount.ID,
			AccountNumber:     fromAccount.AccountNumber,
			AccountName:       fromAccount.AccountName,
			TransactionType:   "D", // Debit (-)
			Amount:            request.Amount,
			TransactionTime:   transactionTime,
			SourceNumber:      fromAccount.AccountNumber,
			BeneficiaryNumber: toAccount.AccountNumber,
		}

		_, err = svc.Service.TransactionRepo.AddTransaction(debitTransaction)
		if err != nil {
			return err
		}

		// Tambah saldo penerima dengan operator "+"
		lastBalance, err = svc.Service.AccountRepo.IncrementDecrementLastBalance(
			toAccount.ID,
			request.Amount,
			"+", // Credit operator
			updatedAt,
			tx,
		)
		if err != nil {
			return err
		}
		toBalanceAfter = lastBalance

		// Record transaksi Credit untuk penerima
		creditTransaction := models.Transaction{
			AccountID:         toAccount.ID,
			AccountNumber:     toAccount.AccountNumber,
			AccountName:       toAccount.AccountName,
			TransactionType:   "C", // Credit (+)
			Amount:            request.Amount,
			TransactionTime:   transactionTime,
			SourceNumber:      fromAccount.AccountNumber,
			BeneficiaryNumber: toAccount.AccountNumber,
		}

		_, err = svc.Service.TransactionRepo.AddTransaction(creditTransaction)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		// Handle custom transaction errors
		if txErr, ok := err.(*utils.TransactionError); ok {
			if txErr.Code == constans.ACCOUNT_BALANCE_BELOW_MINIMUM_CODE {
				result = helpers.ResponseJSON(false, constans.ACCOUNT_BALANCE_BELOW_MINIMUM_CODE, txErr.Message, nil)
				return ctx.JSON(http.StatusBadRequest, result)
			}
		}

		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, "Transaction failed: "+err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	response := map[string]interface{}{
		"transaction_date": transactionTime,
		"sender": map[string]interface{}{
			"account_number": request.FromAccountNumber,
			"account_name":   fromAccount.AccountName,
			"balance_before": fromBalanceBefore,
			"balance_after":  fromBalanceAfter,
		},
		"beneficiary": map[string]interface{}{
			"account_number": request.ToAccountNumber,
			"account_name":   toAccount.AccountName,
			"balance_before": toBalanceBefore,
			"balance_after":  toBalanceAfter,
		},
		"amount": request.Amount,
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "Transfer successful", response)
	return ctx.JSON(http.StatusOK, result)
}

// checkPINHash verifikasi PIN
func checkPINHash(pin, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pin))
	return err == nil
}
