package transactionService

import (
	"database/sql"
	"fmt"
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
	var (
		result      models.Response
		serviceName = "TransactionService.Deposit"
		// request         models.RequestDeposit
		request = new(models.RequestDeposit)
		account models.Account

		transactionTime             = time.Now()
		updatedAt                   = transactionTime.Format(constans.LAYOUT_TIMESTAMP)
		balanceAfter, balanceBefore float64

		transaction models.Transaction
		response    models.DepositResponse
	)

	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		utils.LogError(serviceName, constans.EMPTY_VALUE, "Deposit.BindValidateStruct", err)
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	utils.LogInfo(serviceName, request.AccountNumber, "Deposit",
		fmt.Sprintf("Request amount: %.2f", request.Amount))

	account, err := svc.Service.AccountRepo.FindAccountByNumber(request.AccountNumber)
	if err != nil {
		utils.LogError(serviceName, request.AccountNumber, "Deposit.FindAccountByNumber", err)
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	// Check account status
	if account.AccountStatus == "BLOCKED_PIN" {
		utils.LogError(serviceName, request.AccountNumber, "Deposit.CheckAccountStatus",
			fmt.Errorf("Account is blocked"))
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Account is blocked. Please reset your PIN", nil)
		return ctx.JSON(http.StatusForbidden, result)
	}

	// Verify PIN with failed attempts tracking
	if !helpers.CheckPINHash(request.PIN, account.PIN) {
		failedAttempts, _ := svc.Service.AccountRepo.IncrementFailedPINAttempts(request.AccountNumber)
		remainingAttempts := 3 - failedAttempts

		utils.LogError(serviceName, request.AccountNumber, "Deposit.VerifyPIN",
			fmt.Errorf("Invalid PIN. Remaining attempts: %d", remainingAttempts))

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
		balanceBefore = account.Balance
		balanceAfter = lastBalance

		transaction = models.Transaction{
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
		utils.LogError(serviceName, request.AccountNumber, "Deposit.DBTransaction", err)
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, "Transaction failed: "+err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	utils.LogInfo(serviceName, request.AccountNumber, "Deposit.Success",
		fmt.Sprintf("Amount: %.2f, Balance Before: %.2f, Balance After: %.2f", request.Amount, balanceBefore, balanceAfter))

	response = models.DepositResponse{
		AccountNumber:   account.AccountNumber,
		AccountName:     account.AccountName,
		BalanceBefore:   balanceBefore,
		Amount:          request.Amount,
		BalanceAfter:    balanceAfter,
		TransactionDate: updatedAt,
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "Deposit successful", response)
	return ctx.JSON(http.StatusOK, result)
}

// Withdraw menarik saldo
func (svc transactionService) Withdraw(ctx echo.Context) error {
	var (
		result      models.Response
		serviceName = "TransactionService.Withdraw"
		// request         models.RequestWithdraw
		request                     = new(models.RequestWithdraw)
		account                     models.Account
		transactionTime             = time.Now()
		updatedAt                   = transactionTime.Format(constans.LAYOUT_TIMESTAMP)
		balanceAfter, balanceBefore float64

		transaction models.Transaction
		response    models.WithdrawResponse
	)

	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		utils.LogError(serviceName, constans.EMPTY_VALUE, "Withdraw.BindValidateStruct", err)
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	utils.LogInfo(serviceName, request.AccountNumber, "Withdraw",
		fmt.Sprintf("Request amount: %.2f", request.Amount))

	account, err := svc.Service.AccountRepo.FindAccountByNumber(request.AccountNumber)
	if err != nil {
		utils.LogError(serviceName, request.AccountNumber, "Withdraw.FindAccountByNumber", err)
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	// Check account status
	if account.AccountStatus == "BLOCKED_PIN" {
		utils.LogError(serviceName, request.AccountNumber, "Withdraw.CheckAccountStatus",
			fmt.Errorf("Account is blocked"))
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Account is blocked. Please reset your PIN", nil)
		return ctx.JSON(http.StatusForbidden, result)
	}

	// Verify PIN with failed attempts tracking
	if !helpers.CheckPINHash(request.PIN, account.PIN) {
		failedAttempts, _ := svc.Service.AccountRepo.IncrementFailedPINAttempts(request.AccountNumber)
		remainingAttempts := 3 - failedAttempts

		utils.LogError(serviceName, request.AccountNumber, "Withdraw.VerifyPIN",
			fmt.Errorf("Invalid PIN. Remaining attempts: %d", remainingAttempts))

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
		utils.LogError(serviceName, request.AccountNumber, "Withdraw.CheckBalance",
			fmt.Errorf("Insufficient balance. Current: %.2f, Requested: %.2f", account.Balance, request.Amount))
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Insufficient balance", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

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
		balanceBefore = account.Balance
		balanceAfter = lastBalance

		if balanceAfter < 0 {
			return &utils.TransactionError{
				Code:    constans.ACCOUNT_BALANCE_BELOW_MINIMUM_CODE,
				Message: "Account balance below minimum",
			}
		}

		transaction = models.Transaction{
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
				utils.LogError(serviceName, request.AccountNumber, "Withdraw.BalanceBelowMinimum",
					fmt.Errorf("%s", txErr.Message))
				result = helpers.ResponseJSON(false, constans.ACCOUNT_BALANCE_BELOW_MINIMUM_CODE, txErr.Message, nil)
				return ctx.JSON(http.StatusBadRequest, result)
			}
		}

		utils.LogError(serviceName, request.AccountNumber, "Withdraw.DBTransaction", err)
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, "Transaction failed: "+err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	utils.LogInfo(serviceName, request.AccountNumber, "Withdraw.Success",
		fmt.Sprintf("Amount: %.2f, Balance Before: %.2f, Balance After: %.2f", request.Amount, balanceBefore, balanceAfter))

	response = models.WithdrawResponse{
		AccountNumber:   account.AccountNumber,
		AccountName:     account.AccountName,
		BalanceBefore:   balanceBefore,
		Amount:          request.Amount,
		BalanceAfter:    balanceAfter,
		TransactionDate: updatedAt,
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "Withdraw successful", response)
	return ctx.JSON(http.StatusOK, result)
}

// Transfer antar akun
func (svc transactionService) Transfer(ctx echo.Context) error {
	var (
		result      models.Response
		serviceName = "TransactionService.Transfer"
		// request          models.RequestTransfer
		request          = new(models.RequestTransfer)
		fromBalanceAfter float64
		toBalanceAfter   float64

		debitTransaction  models.Transaction
		creditTransaction models.Transaction
		response          models.TransferResponse
	)

	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		utils.LogError(serviceName, constans.EMPTY_VALUE, "Transfer.BindValidateStruct", err)
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	utils.LogInfo(serviceName, request.FromAccountNumber, "Transfer",
		fmt.Sprintf("To: %s, Amount: %.2f", request.ToAccountNumber, request.Amount))

	if request.FromAccountNumber == request.ToAccountNumber {
		utils.LogError(serviceName, request.FromAccountNumber, "Transfer.ValidateSameAccount",
			fmt.Errorf("Cannot transfer to same account"))
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Cannot transfer to same account", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	fromAccount, err := svc.Service.AccountRepo.FindAccountByNumber(request.FromAccountNumber)
	if err != nil {
		utils.LogError(serviceName, request.FromAccountNumber, "Transfer.FindSourceAccount", err)
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Source account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	// Check account status
	if fromAccount.AccountStatus == "BLOCKED_PIN" {
		utils.LogError(serviceName, request.FromAccountNumber, "Transfer.CheckSourceAccountStatus",
			fmt.Errorf("Source account is blocked"))
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Account is blocked. Please reset your PIN", nil)
		return ctx.JSON(http.StatusForbidden, result)
	}

	// Verify PIN with failed attempts tracking
	if !helpers.CheckPINHash(request.PIN, fromAccount.PIN) {
		failedAttempts, _ := svc.Service.AccountRepo.IncrementFailedPINAttempts(request.FromAccountNumber)
		remainingAttempts := 3 - failedAttempts

		utils.LogError(serviceName, request.FromAccountNumber, "Transfer.VerifyPIN",
			fmt.Errorf("Invalid PIN. Remaining attempts: %d", remainingAttempts))

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
		utils.LogError(serviceName, request.FromAccountNumber, "Transfer.CheckBalance",
			fmt.Errorf("Insufficient balance. Current: %.2f, Requested: %.2f", fromAccount.Balance, request.Amount))
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Insufficient balance", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	toAccount, err := svc.Service.AccountRepo.FindAccountByNumber(request.ToAccountNumber)
	if err != nil {
		utils.LogError(serviceName, request.FromAccountNumber, "Transfer.FindBeneficiaryAccount", err,
			fmt.Sprintf("Beneficiary account: %s", request.ToAccountNumber))
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Beneficiary account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	fromBalanceBefore := fromAccount.Balance
	toBalanceBefore := toAccount.Balance
	transactionTime := time.Now()
	updatedAt := transactionTime.Format(constans.LAYOUT_TIMESTAMP)

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

		debitTransaction = models.Transaction{
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

		creditTransaction = models.Transaction{
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
				utils.LogError(serviceName, request.FromAccountNumber, "Transfer.BalanceBelowMinimum",
					fmt.Errorf("%s", txErr.Message))
				result = helpers.ResponseJSON(false, constans.ACCOUNT_BALANCE_BELOW_MINIMUM_CODE, txErr.Message, nil)
				return ctx.JSON(http.StatusBadRequest, result)
			}
		}

		utils.LogError(serviceName, request.FromAccountNumber, "Transfer.DBTransaction", err)
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, "Transaction failed: "+err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	utils.LogInfo(serviceName, request.FromAccountNumber, "Transfer.Success",
		fmt.Sprintf("To: %s (%s), Amount: %.2f, From Balance: %.2f->%.2f, To Balance: %.2f->%.2f",
			toAccount.AccountNumber, toAccount.AccountName, request.Amount,
			fromBalanceBefore, fromBalanceAfter, toBalanceBefore, toBalanceAfter))

	response = models.TransferResponse{
		FromAccountNumber: request.FromAccountNumber,
		ToAccountNumber:   request.ToAccountNumber,
		Amount:            request.Amount,
		FromBalanceBefore: fromBalanceBefore,
		FromBalanceAfter:  fromBalanceAfter,
		ToBalanceBefore:   toBalanceBefore,
		ToBalanceAfter:    toBalanceAfter,
		TransactionDate:   time.Now(),
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "Transfer successful", response)
	return ctx.JSON(http.StatusOK, result)
}

// GetTransactionHistory mendapatkan riwayat transaksi
func (svc transactionService) GetTransactionHistory(ctx echo.Context) error {
	var (
		result      models.Response
		serviceName = "TransactionService.GetTransactionHistory"
		// request      models.RequestTransactionHistory
		request      = new(models.RequestTransactionHistory)
		response     models.TransactionHistorySimpleResponse
		responseData models.TransactionListResponse
	)

	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		utils.LogError(serviceName, constans.EMPTY_VALUE, "GetTransactionHistory.BindValidateStruct", err)
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	refNo := request.AccountNumber
	if refNo == "" {
		refNo = "ALL_ACCOUNTS"
	}

	utils.LogInfo(serviceName, refNo, "GetTransactionHistory",
		fmt.Sprintf("StartDate: %s, EndDate: %s, Limit: %d, Page: %d",
			request.StartDate, request.EndDate, request.Limit, request.Page))

	// Jika account number diberikan, verifikasi account exists
	if request.AccountNumber != "" {
		_, err := svc.Service.AccountRepo.FindAccountByNumber(request.AccountNumber)
		if err != nil {
			utils.LogError(serviceName, request.AccountNumber, "GetTransactionHistory.FindAccountByNumber", err)
			result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Account not found", nil)
			return ctx.JSON(http.StatusNotFound, result)
		}
	}

	// Deteksi apakah pagination diminta (limit > 0 atau page > 0)
	usePagination := request.Limit > 0 || request.Page > 0

	var limit, page int
	if usePagination {
		// Set default values untuk pagination
		limit = request.Limit
		page = request.Page

		if limit <= 0 {
			limit = 10
		}
		if page <= 0 {
			page = 1
		}

		// Jika tidak ada account number, set limit maksimal untuk prevent overload
		if request.AccountNumber == "" && limit > 100 {
			limit = 100
			utils.LogInfo(serviceName, refNo, "GetTransactionHistory.LimitAdjusted",
				"Limit adjusted to 100 for all accounts query")
		}
	} else {
		// Tanpa pagination: set limit dan page ke 0
		limit = 0
		page = 0
	}

	// Get transaction history
	transactions, totalRecords, err := svc.Service.TransactionRepo.GetTransactionHistory(
		request.AccountNumber,
		request.StartDate,
		request.EndDate,
		limit,
		page,
	)

	if err != nil {
		utils.LogError(serviceName, refNo, "GetTransactionHistory.GetTransactionHistory", err)
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	// Convert to simple response format
	var transactionResponses []models.TransactionSimpleResponse
	for _, tx := range transactions {
		transactionResponses = append(transactionResponses, tx.ToSimpleResponse())
	}

	utils.LogInfo(serviceName, refNo, "GetTransactionHistory.Success",
		fmt.Sprintf("Retrieved %d transactions (Total: %d)", len(transactionResponses), totalRecords))

	message := "Transaction history retrieved successfully"
	if request.AccountNumber == "" {
		message = "All transactions retrieved successfully"
	}

	// Response berbeda tergantung apakah ada pagination atau tidak
	if usePagination {
		// Dengan pagination
		totalPages := int(math.Ceil(float64(totalRecords) / float64(limit)))

		response = models.TransactionHistorySimpleResponse{
			Transactions: transactionResponses,
			Pagination: models.PaginationMeta{
				CurrentPage:  page,
				PerPage:      limit,
				TotalRecords: totalRecords,
				TotalPages:   totalPages,
			},
		}

		result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, message, response)
	} else {
		// Tanpa pagination - hanya return array transactions
		responseData = models.TransactionListResponse{
			Transactions: make([]models.TransactionResponse, 0, len(transactionResponses)),
			TotalRecords: totalRecords,
		}

		for _, tx := range transactionResponses {
			responseData.Transactions = append(responseData.Transactions, models.TransactionResponse{
				ID:            tx.ID,
				AccountNumber: tx.AccountNumber,
				Amount:        tx.Amount,
				// TransactionDate: tx.TransactionDate.Format(time.RFC3339),
				// Description:   tx.Description,
			})
		}

		result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, message, responseData)
	}

	return ctx.JSON(http.StatusOK, result)
}

// GetTransactionDetail mendapatkan detail transaksi
func (svc transactionService) GetTransactionDetail(ctx echo.Context) error {
	var (
		result      models.Response
		serviceName = "TransactionService.GetTransactionDetail"
		// request     models.RequestTransactionDetail
		request  = new(models.RequestTransactionDetail)
		response models.TransactionDetailResponse
	)

	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		utils.LogError(serviceName, constans.EMPTY_VALUE, "GetTransactionDetail.BindValidateStruct", err)
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	refNo := fmt.Sprintf("TRX_ID_%d", request.TransactionID)
	utils.LogInfo(serviceName, refNo, "GetTransactionDetail", "Request received")

	// Get transaction by ID
	transaction, err := svc.Service.TransactionRepo.FindTransactionById(request.TransactionID)
	if err != nil {
		utils.LogError(serviceName, refNo, "GetTransactionDetail.FindTransactionById", err)
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Transaction not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	utils.LogInfo(serviceName, refNo, "GetTransactionDetail.Success",
		fmt.Sprintf("Account: %s, Type: %s, Amount: %.2f",
			transaction.AccountNumber, transaction.TransactionType, transaction.Amount))

	response = models.TransactionDetailResponse{
		ID:                transaction.ID,
		AccountNumber:     transaction.AccountNumber,
		AccountName:       transaction.AccountName,
		SourceNumber:      transaction.SourceNumber,
		BeneficiaryNumber: transaction.BeneficiaryNumber,
		TransactionType:   transaction.TransactionType,
		Amount:            transaction.Amount,
		CreatedAt:         time.Now().Format(constans.LAYOUT_TIMESTAMP),
		TransactionTime:   transaction.TransactionTime.Format(constans.LAYOUT_TIMESTAMP),
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "Transaction detail retrieved successfully", response)
	return ctx.JSON(http.StatusOK, result)
}
