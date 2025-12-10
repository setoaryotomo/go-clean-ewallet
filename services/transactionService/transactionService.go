package transactionService

import (
	"database/sql"
	"math"
	"net/http"
	"sample/constans"
	"sample/helpers"
	"sample/models"
	"sample/services"
	"sample/utils"
	"strconv"
	"time"

	"github.com/labstack/echo"
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

	account, err := svc.Service.AccountRepo.FindAccountByNumber(request.AccountNumber)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	balanceBefore := account.Balance
	transactionTime := time.Now()
	updatedAt := transactionTime.Format("2006-01-02 15:04:05")

	var balanceAfter float64

	err = utils.DBTransaction(svc.Service.RepoDB, func(tx *sql.Tx) error {
		lastBalance, err := svc.Service.AccountRepo.IncrementDecrementLastBalance(
			account.ID,
			request.Amount,
			"+",
			updatedAt,
			tx,
		)
		if err != nil {
			return err
		}
		balanceAfter = lastBalance

		transaction := models.Transaction{
			AccountID:         account.ID,
			AccountNumber:     account.AccountNumber,
			AccountName:       account.AccountName,
			TransactionType:   "C",
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

	account, err := svc.Service.AccountRepo.FindAccountByNumber(request.AccountNumber)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	// Check account status
	if account.AccountStatus == "BLOCKED_PIN" {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Account is blocked. Please reset your PIN", nil)
		return ctx.JSON(http.StatusForbidden, result)
	}

	// Verify PIN with failed attempts tracking
	if !helpers.CheckPINHash(request.PIN, account.PIN) {
		failedAttempts, _ := svc.Service.AccountRepo.IncrementFailedPINAttempts(request.AccountNumber)

		remainingAttempts := 3 - failedAttempts
		if remainingAttempts <= 0 {
			result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Account blocked due to multiple failed PIN attempts", nil)
			return ctx.JSON(http.StatusForbidden, result)
		}

		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE,
			"Invalid PIN. "+strconv.Itoa(remainingAttempts)+" attempt(s) remaining", nil)
		return ctx.JSON(http.StatusUnauthorized, result)
	}

	// Reset failed attempts on successful PIN
	svc.Service.AccountRepo.ResetFailedPINAttempts(request.AccountNumber)

	if account.Balance < request.Amount {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Insufficient balance", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	balanceBefore := account.Balance
	transactionTime := time.Now()
	updatedAt := transactionTime.Format("2006-01-02 15:04:05")

	var balanceAfter float64

	err = utils.DBTransaction(svc.Service.RepoDB, func(tx *sql.Tx) error {
		lastBalance, err := svc.Service.AccountRepo.IncrementDecrementLastBalance(
			account.ID,
			request.Amount,
			"-",
			updatedAt,
			tx,
		)
		if err != nil {
			return err
		}
		balanceAfter = lastBalance

		if balanceAfter < 0 {
			return &utils.TransactionError{
				Code:    constans.ACCOUNT_BALANCE_BELOW_MINIMUM_CODE,
				Message: "Account balance below minimum",
			}
		}

		transaction := models.Transaction{
			AccountID:         account.ID,
			AccountNumber:     account.AccountNumber,
			AccountName:       account.AccountName,
			TransactionType:   "D",
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

	if request.FromAccountNumber == request.ToAccountNumber {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Cannot transfer to same account", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	fromAccount, err := svc.Service.AccountRepo.FindAccountByNumber(request.FromAccountNumber)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Source account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	// Check account status
	if fromAccount.AccountStatus == "BLOCKED_PIN" {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Account is blocked. Please reset your PIN", nil)
		return ctx.JSON(http.StatusForbidden, result)
	}

	// Verify PIN with failed attempts tracking
	if !helpers.CheckPINHash(request.PIN, fromAccount.PIN) {
		failedAttempts, _ := svc.Service.AccountRepo.IncrementFailedPINAttempts(request.FromAccountNumber)

		remainingAttempts := 3 - failedAttempts
		if remainingAttempts <= 0 {
			result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Account blocked due to multiple failed PIN attempts", nil)
			return ctx.JSON(http.StatusForbidden, result)
		}

		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE,
			"Invalid PIN. "+strconv.Itoa(remainingAttempts)+" attempt(s) remaining", nil)
		return ctx.JSON(http.StatusUnauthorized, result)
	}

	// Reset failed attempts on successful PIN
	svc.Service.AccountRepo.ResetFailedPINAttempts(request.FromAccountNumber)

	if fromAccount.Balance < request.Amount {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Insufficient balance", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

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

	err = utils.DBTransaction(svc.Service.RepoDB, func(tx *sql.Tx) error {
		lastBalance, err := svc.Service.AccountRepo.IncrementDecrementLastBalance(
			fromAccount.ID,
			request.Amount,
			"-",
			updatedAt,
			tx,
		)
		if err != nil {
			return err
		}
		fromBalanceAfter = lastBalance

		if fromBalanceAfter < 0 {
			return &utils.TransactionError{
				Code:    constans.ACCOUNT_BALANCE_BELOW_MINIMUM_CODE,
				Message: "Sender balance would be negative after transfer",
			}
		}

		debitTransaction := models.Transaction{
			AccountID:         fromAccount.ID,
			AccountNumber:     fromAccount.AccountNumber,
			AccountName:       fromAccount.AccountName,
			TransactionType:   "D",
			Amount:            request.Amount,
			TransactionTime:   transactionTime,
			SourceNumber:      fromAccount.AccountNumber,
			BeneficiaryNumber: toAccount.AccountNumber,
		}

		_, err = svc.Service.TransactionRepo.AddTransaction(debitTransaction)
		if err != nil {
			return err
		}

		lastBalance, err = svc.Service.AccountRepo.IncrementDecrementLastBalance(
			toAccount.ID,
			request.Amount,
			"+",
			updatedAt,
			tx,
		)
		if err != nil {
			return err
		}
		toBalanceAfter = lastBalance

		creditTransaction := models.Transaction{
			AccountID:         toAccount.ID,
			AccountNumber:     toAccount.AccountNumber,
			AccountName:       toAccount.AccountName,
			TransactionType:   "C",
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

// GetTransactionHistory mendapatkan riwayat transaksi (account number optional)
func (svc transactionService) GetTransactionHistory(ctx echo.Context) error {
	var result models.Response

	request := new(models.RequestTransactionHistory)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Jika account number diberikan, verifikasi account exists
	if request.AccountNumber != "" {
		_, err := svc.Service.AccountRepo.FindAccountByNumber(request.AccountNumber)
		if err != nil {
			result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Account not found", nil)
			return ctx.JSON(http.StatusNotFound, result)
		}
	}

	// Set default values
	if request.Limit <= 0 {
		request.Limit = 10
	}
	if request.Page <= 0 {
		request.Page = 1
	}

	// Jika tidak ada account number, set limit maksimal untuk prevent overload
	if request.AccountNumber == "" && request.Limit > 100 {
		request.Limit = 100
	}

	// Get transaction history
	transactions, totalRecords, err := svc.Service.TransactionRepo.GetTransactionHistory(
		request.AccountNumber,
		request.StartDate,
		request.EndDate,
		request.Limit,
		request.Page,
	)

	if err != nil {
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	// Convert to simple response format
	var transactionResponses []models.TransactionSimpleResponse
	for _, tx := range transactions {
		transactionResponses = append(transactionResponses, tx.ToSimpleResponse())
	}

	// Calculate pagination
	totalPages := int(math.Ceil(float64(totalRecords) / float64(request.Limit)))

	response := models.TransactionHistorySimpleResponse{
		Transactions: transactionResponses,
		Pagination: models.PaginationMeta{
			CurrentPage:  request.Page,
			PerPage:      request.Limit,
			TotalRecords: totalRecords,
			TotalPages:   totalPages,
		},
	}

	message := "Transaction history retrieved successfully"
	if request.AccountNumber == "" {
		message = "All transactions retrieved successfully"
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, message, response)
	return ctx.JSON(http.StatusOK, result)
}

// GetTransactionDetail mendapatkan detail transaksi
func (svc transactionService) GetTransactionDetail(ctx echo.Context) error {
	var result models.Response

	request := new(models.RequestTransactionDetail)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Get transaction by ID
	transaction, err := svc.Service.TransactionRepo.FindTransactionById(request.TransactionID)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Transaction not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	response := transaction.ToDetailResponse()

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "Transaction detail retrieved successfully", response)
	return ctx.JSON(http.StatusOK, result)
}

// checkPINHash verifikasi PIN
// func checkPINHash(pin, hash string) bool {
// 	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pin))
// 	return err == nil
// }
