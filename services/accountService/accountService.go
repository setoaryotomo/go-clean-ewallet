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

	request := new(models.RequestCreateAccount)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	if len(request.PIN) != 6 {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "panjang PIN harus 6 karakter", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	if !helpers.IsNumeric(request.PIN) {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "PIN harus berupa angka", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	hashedPIN, err := helpers.HashPIN(request.PIN)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, "Failed to hash PIN", nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	if request.InitialDeposit < 0 {
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
			result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, "Failed to generate unique account number", nil)
			return ctx.JSON(http.StatusInternalServerError, result)
		}
	}

	tx, err := svc.Service.RepoDB.Begin()
	if err != nil {
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
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	if err := tx.Commit(); err != nil {
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, "Failed to commit transaction", nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	response := models.AccountResponse{
		ID:            id,
		AccountNumber: account.AccountNumber,
		AccountName:   account.AccountName,
		Balance:       account.Balance,
		AccountStatus: "ACTIVE",
		CreatedAt:     account.CreatedAt,
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "Account created successfully", response)
	return ctx.JSON(http.StatusOK, result)
}

// ChangePIN ubah PIN akun dengan satu DBTransaction
func (svc accountService) ChangePIN(ctx echo.Context) error {
	var result models.Response

	request := new(models.RequestChangePIN)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Validasi format PIN baru
	if !helpers.IsNumeric(request.NewPIN) {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "PIN harus berupa angka", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Validasi PIN lama tidak sama dengan PIN baru
	if request.OldPIN == request.NewPIN {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "PIN baru tidak boleh sama dengan PIN lama", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Cek akun exists dan ambil current hashed PIN
	account, err := svc.Service.AccountRepo.FindAccountByNumber(request.AccountNumber)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	// Cek status akun
	if account.AccountStatus == "BLOCKED_PIN" {
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
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	// Sukses
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
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	account, err := svc.Service.AccountRepo.FindAccountByNumber(request.AccountNumber)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	resetToken := helpers.GenerateResetToken()

	//set expiry time
	expiryTime := time.Now().Add(5 * time.Minute)
	err = config.SetResetToken(resetToken, account.AccountNumber, expiryTime)
	if err != nil {
		helpers.LOG("Failed to store reset token", err, false)
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE,
			"Failed to generate reset token. Please try again", nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

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

	request := new(models.RequestResetPIN)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Validasi format PIN baru
	if !helpers.IsNumeric(request.NewPIN) {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "PIN must be numeric", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Validasi PIN confirmation
	if request.NewPIN != request.ConfirmNewPIN {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "New PIN and Confirm PIN do not match", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Dapatkan account_number dari token
	accountNumber, err := config.GetAccountNumberByToken(request.ResetToken)
	if err != nil {
		if helpers.Contains(err.Error(), "expired or not found") {
			result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE,
				"Reset token has expired or is invalid. Please request a new token", nil)
			return ctx.JSON(http.StatusBadRequest, result)
		}
		helpers.LOG("Failed to verify reset token", err, false)
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE,
			"Failed to verify reset token", nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	// Verifikasi account masih exists
	account, err := svc.Service.AccountRepo.FindAccountByNumber(accountNumber)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	// Hash PIN baru
	hashedPIN, err := helpers.HashPIN(request.NewPIN)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, "Failed to process new PIN", nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	// Update PIN menggunakan transaction
	err = utils.DBTransaction(svc.Service.RepoDB, func(tx *sql.Tx) error {
		return svc.Service.AccountRepo.UpdatePINWithTx(tx, accountNumber, hashedPIN)
	})

	if err != nil {
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE,
			"Failed to reset PIN: "+err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	// Hapus token dari Redis setelah berhasil digunakan
	if err := config.DeleteResetToken(request.ResetToken); err != nil {
		helpers.LOG("Failed to delete reset token from Redis", err, false)
		// Tidak perlu gagalkan request, token akan expire sendiri
	}

	response := models.ResetPINResponse{
		AccountNumber: account.AccountNumber,
		ResetAt:       time.Now(),
		// Message:       "PIN has been reset successfully. You can now use your new PIN",
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "PIN reset successfully", response)
	return ctx.JSON(http.StatusOK, result)
}

// GetAccountList mendapatkan list semua akun
func (svc accountService) GetAccountList(ctx echo.Context) error {
	var result models.Response

	accounts, err := svc.Service.AccountRepo.GetAccountList()
	if err != nil {
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	if len(accounts) == 0 {
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
			CreatedAt:     acc.CreatedAt,
		})
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "Account retrieved successfully", accountResponses)
	return ctx.JSON(http.StatusOK, result)
}

// GetAccountByID mendapatkan detail akun berdasarkan ID
func (svc accountService) GetAccountByID(ctx echo.Context) error {
	var result models.Response

	request := new(models.RequestGetAccountByID)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	account, err := svc.Service.AccountRepo.FindAccountById(request.ID)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	response := models.AccountResponse{
		ID:            account.ID,
		AccountNumber: account.AccountNumber,
		AccountName:   account.AccountName,
		Balance:       account.Balance,
		AccountStatus: account.AccountStatus,
		CreatedAt:     account.CreatedAt,
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "Account retrieved successfully", response)
	return ctx.JSON(http.StatusOK, result)
}

// UpdateAccount update data akun
func (svc accountService) UpdateAccount(ctx echo.Context) error {
	var result models.Response

	request := new(models.RequestUpdateAccount)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	_, err := svc.Service.AccountRepo.FindAccountById(request.ID)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	account := models.Account{
		ID:          request.ID,
		AccountName: request.AccountName,
	}

	id, err := svc.Service.AccountRepo.UpdateAccount(account)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "Account updated successfully", id)
	return ctx.JSON(http.StatusOK, result)
}

// DeleteAccount delete akun
func (svc accountService) DeleteAccount(ctx echo.Context) error {
	var result models.Response

	request := new(models.RequestDeleteAccount)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	account, err := svc.Service.AccountRepo.FindAccountById(request.ID)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	if account.Balance > 0 {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Cannot delete account with remaining balance", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	err = svc.Service.AccountRepo.RemoveAccount(request.ID)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "Account deleted successfully", map[string]interface{}{
		"deleted_account_id": request.ID,
		"account_number":     account.AccountNumber,
		"account_name":       account.AccountName,
	})
	return ctx.JSON(http.StatusOK, result)
}
