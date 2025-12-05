package accountService

import (
	"fmt"
	"math/rand"
	"net/http"
	"sample/constans"
	"sample/helpers"
	"sample/models"
	"sample/services"
	"strconv"
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

	// Bind dan validasi request
	request := new(models.RequestCreateAccount)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Hash PIN menggunakan bcrypt
	hashedPIN, err := hashPIN(request.PIN)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, "Failed to hash PIN", nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	// Validasi initial deposit minimum
	if request.InitialDeposit < 0 {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Initial deposit cant negative", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Generate account number otomatis
	accountNumber := generateAccountNumber()

	// Pastikan account number unik (max 5 attempts)
	maxAttempts := 5
	for attempts := 0; attempts < maxAttempts; attempts++ {
		_, exists := svc.Service.AccountRepo.IsAccountExistsByNumber(accountNumber)
		if !exists {
			break
		}
		// Jika sudah ada, generate ulang
		accountNumber = generateAccountNumber()

		if attempts == maxAttempts-1 {
			result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, "Failed to generate unique account number", nil)
			return ctx.JSON(http.StatusInternalServerError, result)
		}
	}

	// Set account data
	account := models.Account{
		AccountNumber: accountNumber,
		AccountName:   request.AccountName,
		Balance:       request.InitialDeposit,
		PIN:           hashedPIN,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Insert ke database
	id, err := svc.Service.AccountRepo.AddAccount(account)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	// Response tanpa menampilkan PIN
	response := models.AccountResponse{
		ID:            id,
		AccountNumber: account.AccountNumber,
		AccountName:   account.AccountName,
		Balance:       account.Balance,
		CreatedAt:     account.CreatedAt,
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "Account created successfully", response)
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

	// Handle jika tidak ada data
	if len(accounts) == 0 {
		result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "No accounts found", []models.AccountResponse{})
		return ctx.JSON(http.StatusOK, result)
	}

	// Convert ke response tanpa PIN
	var accountResponses []models.AccountResponse
	for _, acc := range accounts {
		accountResponses = append(accountResponses, models.AccountResponse{
			ID:            acc.ID,
			AccountNumber: acc.AccountNumber,
			AccountName:   acc.AccountName,
			Balance:       acc.Balance,
			CreatedAt:     acc.CreatedAt,
		})
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, constans.EMPTY_VALUE, accountResponses)
	return ctx.JSON(http.StatusOK, result)
}

// GetAccountByID mendapatkan detail akun berdasarkan ID (POST method)
func (svc accountService) GetAccountByID(ctx echo.Context) error {
	var result models.Response

	// Bind dan validasi request
	request := new(models.RequestGetAccountByID)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Cek account exists by ID
	account, err := svc.Service.AccountRepo.FindAccountById(request.ID)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	// Response tanpa PIN
	response := models.AccountResponse{
		ID:            account.ID,
		AccountNumber: account.AccountNumber,
		AccountName:   account.AccountName,
		Balance:       account.Balance,
		CreatedAt:     account.CreatedAt,
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, constans.EMPTY_VALUE, response)
	return ctx.JSON(http.StatusOK, result)
}

// GetAccountDetailByID mendapatkan detail akun berdasarkan ID (GET method)
func (svc accountService) GetAccountDetailByID(ctx echo.Context) error {
	var result models.Response

	// Get ID from URL parameter
	ID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Invalid account ID", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Cek account exists by ID
	account, err := svc.Service.AccountRepo.FindAccountById(ID)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	// Response tanpa PIN
	response := models.AccountDetailResponse{
		ID:            account.ID,
		AccountNumber: account.AccountNumber,
		AccountName:   account.AccountName,
		Balance:       account.Balance,
		CreatedAt:     account.CreatedAt,
		UpdatedAt:     account.UpdatedAt,
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, constans.EMPTY_VALUE, response)
	return ctx.JSON(http.StatusOK, result)
}

// GetAccountByNumber mendapatkan detail akun berdasarkan nomor akun
func (svc accountService) GetAccountByNumber(ctx echo.Context) error {
	var result models.Response

	accountNumber := ctx.Param("account_number")

	account, err := svc.Service.AccountRepo.FindAccountByNumber(accountNumber)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	// Response tanpa PIN
	response := models.AccountResponse{
		ID:            account.ID,
		AccountNumber: account.AccountNumber,
		AccountName:   account.AccountName,
		Balance:       account.Balance,
		CreatedAt:     account.CreatedAt,
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, constans.EMPTY_VALUE, response)
	return ctx.JSON(http.StatusOK, result)
}

// CheckBalance cek saldo dengan verifikasi PIN
func (svc accountService) CheckBalance(ctx echo.Context) error {
	var result models.Response

	request := new(models.RequestCheckBalance)
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

	response := models.BalanceResponse{
		AccountNumber: account.AccountNumber,
		AccountName:   account.AccountName,
		Balance:       account.Balance,
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, constans.EMPTY_VALUE, response)
	return ctx.JSON(http.StatusOK, result)
}

// Deposit menambah saldo
func (svc accountService) Deposit(ctx echo.Context) error {
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

	// Update balance
	err = svc.Service.AccountRepo.UpdateBalance(request.AccountNumber, request.Amount)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	response := map[string]interface{}{
		"account_number":   account.AccountNumber,
		"account_name":     account.AccountName,
		"balance_before":   account.Balance,
		"deposit_amount":   request.Amount,
		"balance_after":    account.Balance + request.Amount,
		"transaction_date": time.Now(),
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "Deposit successful", response)
	return ctx.JSON(http.StatusOK, result)
}

// Withdraw menarik saldo
func (svc accountService) Withdraw(ctx echo.Context) error {
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

	// Update balance (amount negatif untuk withdraw)
	err = svc.Service.AccountRepo.UpdateBalance(request.AccountNumber, -request.Amount)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	response := map[string]interface{}{
		"account_number":   account.AccountNumber,
		"account_name":     account.AccountName,
		"balance_before":   account.Balance,
		"withdraw_amount":  request.Amount,
		"balance_after":    account.Balance - request.Amount,
		"transaction_date": time.Now(),
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "Withdraw successful", response)
	return ctx.JSON(http.StatusOK, result)
}

// Transfer transfer antar akun
func (svc accountService) Transfer(ctx echo.Context) error {
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
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "From account not found", nil)
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
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "To account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	if toAccount.Balance < 0 {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "error", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Mulai transaction
	tx, err := svc.Service.RepoDB.Begin()
	if err != nil {
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, "Failed to start transaction", nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	// Kurangi saldo pengirim
	err = svc.Service.AccountRepo.UpdateBalance(request.FromAccountNumber, -request.Amount)
	if err != nil {
		tx.Rollback()
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, "Failed to deduct from sender", nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	// Tambah saldo penerima
	err = svc.Service.AccountRepo.UpdateBalance(request.ToAccountNumber, request.Amount)
	if err != nil {
		tx.Rollback()
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, "Failed to add to receiver", nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, "Failed to commit transaction", nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	// Generate transaction ID
	transactionID := fmt.Sprintf("TRF-%s-%d", request.FromAccountNumber, time.Now().Unix())

	response := models.TransferResponse{
		TransactionID:     transactionID,
		FromAccountNumber: request.FromAccountNumber,
		ToAccountNumber:   request.ToAccountNumber,
		Amount:            request.Amount,
		TransactionDate:   time.Now(),
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "Transfer successful", response)
	return ctx.JSON(http.StatusOK, result)
}

// UpdatePIN update PIN akun
func (svc accountService) UpdatePIN(ctx echo.Context) error {
	var result models.Response

	request := new(models.RequestUpdatePIN)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Get account
	account, err := svc.Service.AccountRepo.FindAccountByNumber(request.AccountNumber)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	// Verifikasi old PIN
	if !checkPINHash(request.OldPIN, account.PIN) {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Invalid old PIN", nil)
		return ctx.JSON(http.StatusUnauthorized, result)
	}

	// Hash new PIN
	hashedNewPIN, err := hashPIN(request.NewPIN)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, "Failed to hash new PIN", nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	// Update PIN
	err = svc.Service.AccountRepo.UpdatePIN(request.AccountNumber, hashedNewPIN)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "PIN updated successfully", nil)
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

	// Cek account exists
	_, err := svc.Service.AccountRepo.FindAccountById(request.ID)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	// Set account data
	account := models.Account{
		ID:          request.ID,
		AccountName: request.AccountName,
	}

	// Update
	id, err := svc.Service.AccountRepo.UpdateAccount(account)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "Account updated successfully", id)
	return ctx.JSON(http.StatusOK, result)
}

// DeleteAccount soft delete akun (POST method)
func (svc accountService) DeleteAccount(ctx echo.Context) error {
	var result models.Response

	// Bind dan validasi request
	request := new(models.RequestDeleteAccount)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Cek account exists
	account, err := svc.Service.AccountRepo.FindAccountById(request.ID)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Account not found", nil)
		return ctx.JSON(http.StatusNotFound, result)
	}

	// Validasi: tidak bisa delete jika masih ada saldo
	if account.Balance > 0 {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Cannot delete account with remaining balance", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Soft delete
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

// hashPIN hash PIN menggunakan bcrypt
func hashPIN(pin string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pin), 4)
	return string(bytes), err
}

// checkPINHash verifikasi PIN
func checkPINHash(pin, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pin))
	return err == nil
}

// generateAccountNumber menghasilkan nomor akun unik otomatis
func generateAccountNumber() string {
	// Format: timestamp + random 4 digit
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// Ambil timestamp dalam milidetik
	timestamp := time.Now().UnixNano() / 1000000

	// Generate random 4 digit
	random := rand.Intn(10000)

	// Gabungkan dan ambil 10 digit terakhir
	number := fmt.Sprintf("%d%04d", timestamp, random)

	// Pastikan panjang 10 digit
	if len(number) > 10 {
		number = number[len(number)-10:]
	} else if len(number) < 10 {
		// Tambahkan leading zeros jika kurang dari 10 digit
		number = fmt.Sprintf("%010s", number)
	}

	return number
}
