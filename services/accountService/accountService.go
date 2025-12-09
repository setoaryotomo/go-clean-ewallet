package accountService

import (
	// "crypto/rand"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"sample/constans"
	"sample/helpers"
	"sample/models"
	"sample/services"
	"time"

	"github.com/labstack/echo"
	"golang.org/x/crypto/bcrypt"
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

	if !isNumeric(request.PIN) {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "PIN harus berupa angka", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	hashedPIN, err := hashPIN(request.PIN)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, "Failed to hash PIN", nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	if request.InitialDeposit < 0 {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Initial deposit cant negative", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	accountNumber := generateAccountNumber()

	maxAttempts := 5
	for attempts := 0; attempts < maxAttempts; attempts++ {
		_, exists := svc.Service.AccountRepo.IsAccountExistsByNumber(accountNumber)
		if !exists {
			break
		}
		accountNumber = generateAccountNumber()

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

// ChangePIN ubah PIN akun
func (svc accountService) ChangePIN(ctx echo.Context) error {
	var result models.Response

	request := new(models.RequestChangePIN)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Validasi format PIN baru
	if !isNumeric(request.NewPIN) {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "PIN harus berupa angka", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Validasi PIN lama tidak sama dengan PIN baru
	if request.OldPIN == request.NewPIN {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "PIN baru tidak boleh sama dengan PIN lama", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Cek akun exists
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

	// Verifikasi PIN lama
	if !checkPINHash(request.OldPIN, account.PIN) {
		// Increment failed attempts
		failedAttempts, _ := svc.Service.AccountRepo.IncrementFailedPINAttempts(request.AccountNumber)

		remainingAttempts := 3 - failedAttempts
		if remainingAttempts <= 0 {
			result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Account blocked due to multiple failed PIN attempts", nil)
			return ctx.JSON(http.StatusForbidden, result)
		}

		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE,
			fmt.Sprintf("Invalid old PIN. %d attempt(s) remaining", remainingAttempts), nil)
		return ctx.JSON(http.StatusUnauthorized, result)
	}

	// Hash PIN baru
	hashedPIN, err := hashPIN(request.NewPIN)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, "Failed to hash new PIN", nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	// Update PIN
	err = svc.Service.AccountRepo.UpdatePIN(request.AccountNumber, hashedPIN)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	response := models.ChangePINResponse{
		AccountNumber: request.AccountNumber,
		// Message:       "PIN changed successfully",
		ChangedAt: time.Now(),
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "PIN changed successfully", response)
	return ctx.JSON(http.StatusOK, result)
}

// ForgotPIN untuk reset PIN (simulasi dengan token)
func (svc accountService) ForgotPIN(ctx echo.Context) error {
	var result models.Response

	request := new(models.RequestForgotPIN)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Verifikasi account number dan account name
	account, err := svc.Service.AccountRepo.FindAccountByNumber(request.AccountNumber)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	// Verifikasi account name
	if account.AccountName != request.AccountName {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Account name does not match", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Generate reset token
	resetToken := generateResetToken()

	response := models.ForgotPINResponse{
		AccountNumber: request.AccountNumber,
		// Message:       "Reset token generated. Please use this token to reset your PIN",
		ResetToken: resetToken,
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "Reset token generated successfully", response)
	return ctx.JSON(http.StatusOK, result)
}

// ResetPIN reset PIN dengan token
func (svc accountService) ResetPIN(ctx echo.Context) error {
	var result models.Response

	request := new(models.RequestResetPIN)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Validasi format PIN
	if !isNumeric(request.NewPIN) {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "PIN harus berupa angka", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Validasi reset token
	if len(request.ResetToken) < 32 {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Invalid reset token", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Cek akun exists
	_, err := svc.Service.AccountRepo.FindAccountByNumber(request.AccountNumber)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	// Hash PIN baru
	hashedPIN, err := hashPIN(request.NewPIN)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, "Failed to hash new PIN", nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	// Update PIN dan reset failed attempts
	err = svc.Service.AccountRepo.UpdatePIN(request.AccountNumber, hashedPIN)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	response := map[string]interface{}{
		"account_number": request.AccountNumber,
		// "message":        "PIN reset successfully",
		"reset_at": time.Now(),
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

// Helper functions
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func hashPIN(pin string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pin), 4)
	return string(bytes), err
}

func checkPINHash(pin, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pin))
	return err == nil
}

func generateAccountNumber() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	timestamp := time.Now().UnixNano() / 1000000
	random := rand.Intn(10000)
	number := fmt.Sprintf("%d%04d", timestamp, random)

	if len(number) > 10 {
		number = number[len(number)-10:]
	} else if len(number) < 10 {
		number = fmt.Sprintf("%010s", number)
	}

	return number
}

func generateResetToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
