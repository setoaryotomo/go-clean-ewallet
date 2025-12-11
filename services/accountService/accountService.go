package accountService

import (
	"database/sql"
	"fmt"
	"net/http"
	"sample/config"
	"sample/constans"
	"sample/helpers"
	"sample/models"
	"sample/services"
	"sample/utils"
	"time"

	"github.com/labstack/echo"
)

type accountService struct {
	Service services.UsecaseService
}

// NewAccountService
func NewAccountService(service services.UsecaseService) accountService {
	return accountService{
		Service: service,
	}
}

// CreateAccount membuat akun baru
func (svc accountService) CreateAccount(ctx echo.Context) error {
	var result models.Response
	const serviceName = "CreateAccount"
	request := new(models.RequestCreateAccount)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		utils.LogError(serviceName, constans.EMPTY_VALUE, "CreateAccount.BindValidateStruct", err)
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	utils.LogInfo(serviceName, constans.EMPTY_VALUE, "CreateAccount", fmt.Sprintf("Request: %+v", request))

	if len(request.PIN) != 6 {
		utils.LogError(serviceName, constans.EMPTY_VALUE, "CreateAccount.ValidatePINLength",
			fmt.Errorf("panjang PIN harus 6 karakter"))
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "panjang PIN harus 6 karakter", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	if !helpers.IsNumeric(request.PIN) {
		utils.LogError(serviceName, constans.EMPTY_VALUE, "CreateAccount.ValidatePINNumeric",
			fmt.Errorf("PIN harus berupa angka"))
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "PIN harus berupa angka", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	hashedPIN, err := helpers.HashPIN(request.PIN)
	if err != nil {
		utils.LogError(serviceName, constans.EMPTY_VALUE, "CreateAccount.HashPIN", err)
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, "Failed to hash PIN", nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	if request.InitialDeposit < 0 {
		utils.LogError(serviceName, constans.EMPTY_VALUE, "CreateAccount.ValidateInitialDeposit",
			fmt.Errorf("Initial deposit cant negative"))
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Initial deposit cant negative", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	accountNumber := helpers.GenerateAccountNumber()

	maxAttempts := 5
	for attempts := 0; attempts < maxAttempts; attempts++ {
		_, exists := svc.Service.AccountRepo.IsAccountExistsByNumber(accountNumber)
		if !exists {
			break
		}
		accountNumber = helpers.GenerateAccountNumber()

		if attempts == maxAttempts-1 {
			utils.LogError(serviceName, constans.EMPTY_VALUE, "CreateAccount.GenerateAccountNumber",
				fmt.Errorf("Failed to generate unique account number after %d attempts", maxAttempts))
			result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, "Failed to generate unique account number", nil)
			return ctx.JSON(http.StatusInternalServerError, result)
		}
	}

	tx, err := svc.Service.RepoDB.Begin()
	if err != nil {
		utils.LogError(serviceName, constans.EMPTY_VALUE, "CreateAccount.BeginTransaction", err)
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, "Failed to start transaction", nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	account := models.Account{
		AccountNumber: accountNumber,
		AccountName:   request.AccountName,
		Balance:       request.InitialDeposit,
		PIN:           hashedPIN,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	id, err := svc.Service.AccountRepo.AddAccount(account)
	if err != nil {
		tx.Rollback()
		utils.LogError(serviceName, accountNumber, "CreateAccount.AddAccount", err)
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	if err := tx.Commit(); err != nil {
		utils.LogError(serviceName, accountNumber, "CreateAccount.CommitTransaction", err)
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, "Failed to commit transaction", nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	utils.LogInfo(serviceName, accountNumber, "CreateAccount.Success", fmt.Sprintf("Account ID: %d", id))

	response := models.AccountResponse{
		ID:            id,
		AccountNumber: account.AccountNumber,
		Balance:       account.Balance,
		AccountName:   account.AccountName,
		AccountStatus: "ACTIVE",
		CreatedAt:     account.CreatedAt.Format(time.RFC3339),
	}

	// response := account.ToCreateResponse()

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "Account created successfully", response)
	return ctx.JSON(http.StatusOK, result)
}

// ChangePIN ubah PIN akun dengan satu DBTransaction
func (svc accountService) ChangePIN(ctx echo.Context) error {
	var result models.Response
	serviceName := "AccountService" // Define the serviceName variable
	request := new(models.RequestChangePIN)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		utils.LogError(serviceName, constans.EMPTY_VALUE, "ChangePIN.BindValidateStruct", err)
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	utils.LogInfo(serviceName, request.AccountNumber, "ChangePIN", "Request received")

	// Validasi format PIN baru
	if !helpers.IsNumeric(request.NewPIN) {
		utils.LogError(serviceName, request.AccountNumber, "ChangePIN.ValidatePINNumeric",
			fmt.Errorf("PIN harus berupa angka"))
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "PIN harus berupa angka", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Validasi PIN lama tidak sama dengan PIN baru
	if request.OldPIN == request.NewPIN {
		utils.LogError(serviceName, request.AccountNumber, "ChangePIN.ValidatePINSame",
			fmt.Errorf("PIN baru tidak boleh sama dengan PIN lama"))
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "PIN baru tidak boleh sama dengan PIN lama", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Cek akun exists dan ambil current hashed PIN
	account, err := svc.Service.AccountRepo.FindAccountByNumber(request.AccountNumber)
	if err != nil {
		utils.LogError(serviceName, request.AccountNumber, "ChangePIN.FindAccountByNumber", err)
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	// Cek status akun
	if account.AccountStatus == "BLOCKED_PIN" {
		utils.LogError(serviceName, request.AccountNumber, "ChangePIN.CheckAccountStatus",
			fmt.Errorf("Account is blocked"))
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Account is blocked due to multiple failed PIN attempts. Please use Forgot PIN", nil)
		return ctx.JSON(http.StatusForbidden, result)
	}

	// Gunakan satu DBTransaction untuk semua operasi ChangePIN
	var failedAttempts int
	err = utils.DBTransaction(svc.Service.RepoDB, func(tx *sql.Tx) error {
		// Panggil repository function yang menangani semua operasi dalam satu transaksi
		failedAttempts, err = svc.Service.AccountRepo.ChangePINWithTx(
			tx,
			request.AccountNumber,
			request.OldPIN,
			request.NewPIN,
			account.PIN,
		)
		return err
	})

	if err != nil {
		// Cek jika error adalah PIN verification failed
		if err.Error() == "PIN_VERIFICATION_FAILED" {
			remainingAttempts := 3 - failedAttempts

			utils.LogError(serviceName, request.AccountNumber, "ChangePIN.VerifyPIN",
				fmt.Errorf("Invalid old PIN. Remaining attempts: %d", remainingAttempts))

			if remainingAttempts <= 0 {
				result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE,
					"Account blocked due to multiple failed PIN attempts", nil)
				return ctx.JSON(http.StatusForbidden, result)
			}

			result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE,
				fmt.Sprintf("Invalid old PIN. %d attempt(s) remaining", remainingAttempts), nil)
			return ctx.JSON(http.StatusUnauthorized, result)
		}

		// Error lainnya

		utils.LogError(serviceName, request.AccountNumber, "ChangePIN.Transaction", err)
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	// Sukses
	utils.LogInfo(serviceName, request.AccountNumber, "ChangePIN.Success", "PIN changed successfully")

	response := models.ChangePINResponse{
		AccountNumber: request.AccountNumber,
		ChangedAt:     time.Now(),
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "PIN changed successfully", response)
	return ctx.JSON(http.StatusOK, result)
}

// ForgotPIN Inquiry
func (svc accountService) ForgotPIN(ctx echo.Context) error {
	var result models.Response

	request := new(models.RequestForgotPIN)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		helpers.LOG("ERROR ForgotPIN - Validation failed", err.Error(), false)
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	helpers.LOG("INFO ForgotPIN - Request received", request.AccountNumber, false)

	account, err := svc.Service.AccountRepo.FindAccountByNumber(request.AccountNumber)
	if err != nil {
		helpers.LOG("ERROR ForgotPIN - Account not found", map[string]interface{}{
			"error":          err.Error(),
			"account_number": request.AccountNumber,
		}, false)
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	resetToken := helpers.GenerateResetToken()

	//set expiry time
	expiryTime := time.Now().Add(5 * time.Minute)
	err = config.SetResetToken(resetToken, account.AccountNumber, expiryTime)
	if err != nil {
		helpers.LOG("ERROR ForgotPIN - Failed to store reset token", map[string]interface{}{
			"error":          err.Error(),
			"account_number": account.AccountNumber,
		}, false)
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE,
			"Failed to generate reset token. Please try again", nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	helpers.LOG("SUCCESS ForgotPIN - Reset token generated", map[string]interface{}{
		"account_number": account.AccountNumber,
		// "expires_at":     expiryTime.Format("2006-01-02 15:04:05"),
		"expires_at": expiryTime.Format(constans.LAYOUT_TIMESTAMP),
	}, false)

	response := models.ForgotPINResponse{
		AccountNumber: account.AccountNumber,
		ResetToken:    resetToken,
		ExpiresAt:     expiryTime,
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE,
		"Reset token generated successfully. Please use this token within 5 minutes", response)
	return ctx.JSON(http.StatusOK, result)
}

// ResetPIN reset PIN dengan token dari Redis (Forgot PIN Confirm)
func (svc accountService) ResetPIN(ctx echo.Context) error {
	var result models.Response
	serviceName := "AccountService" // Define the serviceName variable
	request := new(models.RequestResetPIN)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		utils.LogError(serviceName, constans.EMPTY_VALUE, "ResetPIN.BindValidateStruct", err)
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	utils.LogInfo(serviceName, constans.EMPTY_VALUE, "ResetPIN", "Request received")

	// Validasi format PIN baru
	if !helpers.IsNumeric(request.NewPIN) {
		utils.LogError(serviceName, constans.EMPTY_VALUE, "ResetPIN.ValidatePINNumeric",
			fmt.Errorf("PIN must be numeric"))
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "PIN must be numeric", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Validasi PIN confirmation
	if request.NewPIN != request.ConfirmNewPIN {
		utils.LogError(serviceName, constans.EMPTY_VALUE, "ResetPIN.ValidatePINMatch",
			fmt.Errorf("New PIN and Confirm PIN do not match"))
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "New PIN and Confirm PIN do not match", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Dapatkan account_number dari token
	accountNumber, err := config.GetAccountNumberByToken(request.ResetToken)
	if err != nil {
		if helpers.Contains(err.Error(), "expired or not found") {
			utils.LogError(serviceName, constans.EMPTY_VALUE, "ResetPIN.GetAccountNumberByToken", err)
			result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE,
				"Reset token has expired or is invalid. Please request a new token", nil)
			return ctx.JSON(http.StatusBadRequest, result)
		}
		utils.LogError(serviceName, constans.EMPTY_VALUE, "ResetPIN.GetAccountNumberByToken", err)
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE,
			"Failed to verify reset token", nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	// Verifikasi account masih exists
	account, err := svc.Service.AccountRepo.FindAccountByNumber(accountNumber)
	if err != nil {
		utils.LogError(serviceName, accountNumber, "ResetPIN.FindAccountByNumber", err)
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	// Hash PIN baru
	hashedPIN, err := helpers.HashPIN(request.NewPIN)
	if err != nil {
		utils.LogError(serviceName, accountNumber, "ResetPIN.HashPIN", err)
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, "Failed to process new PIN", nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	// Update PIN menggunakan transaction
	err = utils.DBTransaction(svc.Service.RepoDB, func(tx *sql.Tx) error {
		return svc.Service.AccountRepo.UpdatePINWithTx(tx, accountNumber, hashedPIN)
	})

	if err != nil {
		utils.LogError(serviceName, accountNumber, "ResetPIN.UpdatePINWithTx", err)
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE,
			"Failed to reset PIN: "+err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	// Hapus token dari Redis setelah berhasil digunakan
	if err := config.DeleteResetToken(request.ResetToken); err != nil {
		utils.LogError(serviceName, accountNumber, "ResetPIN.DeleteResetToken", err)
		// Tidak perlu gagalkan request, token akan expire sendiri
	}

	utils.LogInfo(serviceName, accountNumber, "ResetPIN.Success", "PIN reset successfully")

	response := models.ResetPINResponse{
		AccountNumber: account.AccountNumber,
		ResetAt:       time.Now(),
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "PIN reset successfully", response)
	return ctx.JSON(http.StatusOK, result)
}

// GetAccountList mendapatkan list semua akun
func (svc accountService) GetAccountList(ctx echo.Context) error {
	var result models.Response

	serviceName := "AccountService" // Initialize serviceName

	utils.LogInfo(serviceName, constans.EMPTY_VALUE, "GetAccountList", "Request received")

	accounts, err := svc.Service.AccountRepo.GetAccountList()
	if err != nil {
		utils.LogError(serviceName, constans.EMPTY_VALUE, "GetAccountList.GetAccountList", err)
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	if len(accounts) == 0 {
		utils.LogInfo(serviceName, constans.EMPTY_VALUE, "GetAccountList", "No accounts found")
		result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "No accounts found", []models.AccountResponse{})
		return ctx.JSON(http.StatusOK, result)
	}

	var accountResponses []models.AccountResponse
	for _, acc := range accounts {
		accountResponses = append(accountResponses, models.AccountResponse{
			ID:            acc.ID,
			AccountNumber: acc.AccountNumber,
			AccountName:   acc.AccountName,
			Balance:       acc.Balance,
			AccountStatus: acc.AccountStatus,
			CreatedAt:     acc.CreatedAt.Format(time.RFC3339), // Convert time.Time to string
		})
	}

	utils.LogInfo(serviceName, constans.EMPTY_VALUE, "GetAccountList.Success",
		fmt.Sprintf("Retrieved %d accounts", len(accountResponses)))

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "Account retrieved successfully", accountResponses)
	return ctx.JSON(http.StatusOK, result)
}

// GetAccountByID mendapatkan detail akun berdasarkan ID
func (svc accountService) GetAccountByID(ctx echo.Context) error {
	var result models.Response
	serviceName := "AccountService"
	request := new(models.RequestGetAccountByID)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		utils.LogError(serviceName, constans.EMPTY_VALUE, "GetAccountByID.BindValidateStruct", err)
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	utils.LogInfo(serviceName, fmt.Sprintf("%d", request.ID), "GetAccountByID", "Request received")

	account, err := svc.Service.AccountRepo.FindAccountById(request.ID)
	if err != nil {
		utils.LogError(serviceName, fmt.Sprintf("%d", request.ID), "GetAccountByID.FindAccountById", err)
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	// response := models.AccountResponse{
	// 	ID:            account.ID,
	// 	AccountNumber: account.AccountNumber,
	// 	AccountName:   account.AccountName,
	// 	Balance:       account.Balance,
	// 	AccountStatus: account.AccountStatus,
	// 	CreatedAt:     account.CreatedAt,
	// }

	response := account.ToAccountResponse()

	utils.LogInfo(serviceName, account.AccountNumber, "GetAccountByID.Success",
		fmt.Sprintf("Account found: %s", account.AccountName))

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "Account retrieved successfully", response)
	return ctx.JSON(http.StatusOK, result)
}

// UpdateAccount update data akun
func (svc accountService) UpdateAccount(ctx echo.Context) error {
	var result models.Response
	serviceName := "AccountService"
	request := new(models.RequestUpdateAccount)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		utils.LogError(serviceName, constans.EMPTY_VALUE, "UpdateAccount.BindValidateStruct", err)
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	utils.LogInfo(serviceName, fmt.Sprintf("%d", request.ID), "UpdateAccount",
		fmt.Sprintf("Request: %+v", request))

	_, err := svc.Service.AccountRepo.FindAccountById(request.ID)
	if err != nil {
		utils.LogError(serviceName, fmt.Sprintf("%d", request.ID), "UpdateAccount.FindAccountById", err)
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	account := models.Account{
		ID:          request.ID,
		AccountName: request.AccountName,
	}

	id, err := svc.Service.AccountRepo.UpdateAccount(account)
	if err != nil {
		utils.LogError(serviceName, fmt.Sprintf("%d", request.ID), "UpdateAccount.UpdateAccount", err)
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	utils.LogInfo(serviceName, fmt.Sprintf("%d", id), "UpdateAccount.Success", "Account updated successfully")

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "Account updated successfully", id)
	return ctx.JSON(http.StatusOK, result)
}

// DeleteAccount delete akun
func (svc accountService) DeleteAccount(ctx echo.Context) error {
	var result models.Response
	serviceName := "AccountService"
	request := new(models.RequestDeleteAccount)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		utils.LogError(serviceName, constans.EMPTY_VALUE, "DeleteAccount.BindValidateStruct", err)
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	utils.LogInfo(serviceName, fmt.Sprintf("%d", request.ID), "DeleteAccount", "Request received")

	account, err := svc.Service.AccountRepo.FindAccountById(request.ID)
	if err != nil {
		utils.LogError(serviceName, fmt.Sprintf("%d", request.ID), "DeleteAccount.FindAccountById", err)
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	if account.Balance > 0 {
		utils.LogError(serviceName, account.AccountNumber, "DeleteAccount.ValidateBalance",
			fmt.Errorf("Cannot delete account with remaining balance: %.2f", account.Balance))
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Cannot delete account with remaining balance", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	err = svc.Service.AccountRepo.RemoveAccount(request.ID)
	if err != nil {
		utils.LogError(serviceName, account.AccountNumber, "DeleteAccount.RemoveAccount", err)
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	utils.LogInfo(serviceName, account.AccountNumber, "DeleteAccount.Success",
		fmt.Sprintf("Account deleted: %s", account.AccountName))

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "Account deleted successfully", map[string]interface{}{
		"deleted_account_id": request.ID,
		"account_number":     account.AccountNumber,
		"account_name":       account.AccountName,
	})
	return ctx.JSON(http.StatusOK, result)
}
