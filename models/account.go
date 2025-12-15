package models

import (
	"sample/constans"
	"time"
)

// Account model
type Account struct {
	ID                int       `json:"id"`
	AccountNumber     string    `json:"account_number"`
	Balance           float64   `json:"balance"`
	PIN               string    `json:"pin,omitempty"`
	AccountName       string    `json:"account_name"`
	AccountStatus     string    `json:"account_status"`
	FailedPINAttempts int       `json:"failed_pin_attempts"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// ============== REQUEST MODELS ==============

type RequestCreateAccount struct {
	AccountName    string  `json:"account_name" validate:"required,min=3,max=255"`
	PIN            string  `json:"pin" validate:"required,len=6"`
	InitialDeposit float64 `json:"initial_deposit" validate:"min=0"`
}

type RequestUpdateAccount struct {
	ID          int    `json:"id" validate:"required,min=1"`
	AccountName string `json:"account_name" validate:"required,min=3,max=255"`
}

type RequestDeleteAccount struct {
	ID int `json:"id" validate:"required,min=1"`
}

type RequestUpdateBalance struct {
	AccountNumber string  `json:"account_number" validate:"required"`
	Amount        float64 `json:"amount" validate:"required,min=10000"`
}

type RequestBalanceInquiry struct {
	AccountNumber string `json:"account_number" validate:"required"`
}

type RequestChangePIN struct {
	AccountNumber string `json:"account_number" validate:"required"`
	OldPIN        string `json:"old_pin" validate:"required,len=6"`
	NewPIN        string `json:"new_pin" validate:"required,len=6"`
}

type RequestForgotPIN struct {
	AccountNumber string `json:"account_number" validate:"required"`
}

type RequestResetPIN struct {
	ResetToken    string `json:"reset_token" validate:"required"`
	NewPIN        string `json:"new_pin" validate:"required,len=6"`
	ConfirmNewPIN string `json:"confirm_new_pin" validate:"required,len=6"`
}

type RequestGetAccountByID struct {
	ID int `json:"id" validate:"required,min=1"`
}

type RequestGetAccountByNumber struct {
	AccountNumber string `json:"account_number" validate:"required"`
}

// ============== RESPONSE MODELS ==============

// BaseAccountResponse - Response dasar untuk account (tanpa balance)
type BaseAccountResponse struct {
	ID            int       `json:"id"`
	AccountNumber string    `json:"account_number"`
	AccountName   string    `json:"account_name"`
	AccountStatus string    `json:"account_status"`
	CreatedAt     time.Time `json:"created_at"`
}

// AccountResponse - Response lengkap dengan balance
type AccountResponse struct {
	ID            int     `json:"id"`
	AccountNumber string  `json:"account_number"`
	Balance       float64 `json:"balance"`
	AccountName   string  `json:"account_name"`
	AccountStatus string  `json:"account_status"`
	// CreatedAt     time.Time `json:"created_at"`
	CreatedAt string `json:"created_at"`
}

type BalanceInquiryResponse struct {
	ID            int     `json:"id"`
	AccountNumber string  `json:"account_number"`
	Balance       float64 `json:"balance"`
	AccountName   string  `json:"account_name"`
	AccountStatus string  `json:"account_status"`
}

// AccountDetailResponse - Response detail dengan UpdatedAt
type AccountDetailResponse struct {
	ID            int     `json:"id"`
	AccountNumber string  `json:"account_number"`
	AccountName   string  `json:"account_name"`
	Balance       float64 `json:"balance"`
	AccountStatus string  `json:"account_status"`
	// CreatedAt     time.Time `json:"created_at"`
	CreatedAt string `json:"created_at"`
	// UpdatedAt     time.Time `json:"updated_at"`
	UpdatedAt string `json:"updated_at"`
}

// CreateAccountResponse - Response khusus untuk create account
type CreateAccountResponse struct {
	ID            int     `json:"id"`
	AccountNumber string  `json:"account_number"`
	AccountName   string  `json:"account_name"`
	Balance       float64 `json:"initial_balance"`
	AccountStatus string  `json:"account_status"`
	// CreatedAt     time.Time `json:"created_at"`
	CreatedAt string `json:"created_at"`
	// Message   string `json:"message"`
}

// UpdateAccountResponse - Response untuk update account
type UpdateAccountResponse struct {
	ID            int    `json:"id"`
	AccountNumber string `json:"account_number"`
	AccountName   string `json:"account_name"`
	// UpdatedAt     time.Time `json:"updated_at"`
	UpdatedAt string `json:"updated_at"`
}

// DeleteAccountResponse - Response untuk delete account
type DeleteAccountResponse struct {
	DeletedAccountID int       `json:"deleted_account_id"`
	AccountNumber    string    `json:"account_number"`
	AccountName      string    `json:"account_name"`
	DeletedAt        time.Time `json:"deleted_at"`
}

// BalanceResponse - Response untuk cek saldo
type BalanceResponse struct {
	AccountNumber string    `json:"account_number"`
	AccountName   string    `json:"account_name"`
	Balance       float64   `json:"balance"`
	CheckedAt     time.Time `json:"checked_at"`
}

// ChangePINResponse - Response untuk change PIN
type ChangePINResponse struct {
	AccountNumber string    `json:"account_number"`
	Message       string    `json:"message"`
	ChangedAt     time.Time `json:"changed_at"`
}

// ForgotPINResponse - Response untuk forgot PIN (generate token)
type ForgotPINResponse struct {
	AccountNumber string    `json:"account_number"`
	ResetToken    string    `json:"reset_token"`
	ExpiresAt     time.Time `json:"expires_at"`
	Message       string    `json:"message"`
}

// ResetPINResponse - Response untuk reset PIN
type ResetPINResponse struct {
	AccountNumber string    `json:"account_number"`
	Message       string    `json:"message"`
	ResetAt       time.Time `json:"reset_at"`
}

// AccountListResponse - Response untuk list accounts
type AccountListResponse struct {
	Accounts     []AccountResponse `json:"accounts"`
	TotalRecords int               `json:"total_records"`
}

// ToBaseResponse converts Account to BaseAccountResponse
func (a *Account) ToBaseResponse() BaseAccountResponse {
	return BaseAccountResponse{
		ID:            a.ID,
		AccountNumber: a.AccountNumber,
		AccountName:   a.AccountName,
		AccountStatus: a.AccountStatus,
		CreatedAt:     a.CreatedAt,
	}
}

// ToAccountResponse converts Account to AccountResponse
func (a *Account) ToAccountResponse() AccountResponse {
	return AccountResponse{
		ID:            a.ID,
		AccountNumber: a.AccountNumber,
		Balance:       a.Balance,
		AccountName:   a.AccountName,
		AccountStatus: a.AccountStatus,
		CreatedAt:     a.CreatedAt.Format(constans.LAYOUT_TIMESTAMP),
	}
}

// ToDetailResponse converts Account to AccountDetailResponse
func (a *Account) ToDetailResponse() AccountDetailResponse {
	return AccountDetailResponse{
		ID:            a.ID,
		AccountNumber: a.AccountNumber,
		AccountName:   a.AccountName,
		Balance:       a.Balance,
		AccountStatus: a.AccountStatus,
		CreatedAt:     a.CreatedAt.Format(constans.LAYOUT_TIMESTAMP),
		UpdatedAt:     a.UpdatedAt.Format(constans.LAYOUT_TIMESTAMP),
	}
}

// ToCreateResponse converts Account to CreateAccountResponse
func (a *Account) ToCreateResponse() CreateAccountResponse {
	return CreateAccountResponse{
		ID:            a.ID,
		AccountNumber: a.AccountNumber,
		AccountName:   a.AccountName,
		Balance:       a.Balance,
		AccountStatus: "ACTIVE",
		CreatedAt:     a.CreatedAt.Format(constans.LAYOUT_TIMESTAMP),
	}
}

// ToUpdateResponse converts Account to UpdateAccountResponse
func (a *Account) ToUpdateResponse() UpdateAccountResponse {
	return UpdateAccountResponse{
		ID:            a.ID,
		AccountNumber: a.AccountNumber,
		AccountName:   a.AccountName,
		UpdatedAt:     a.UpdatedAt.Format(constans.LAYOUT_TIMESTAMP),
	}
}

func ToAccountResponseList(accounts []Account) []AccountResponse {
	var responses []AccountResponse
	for _, acc := range accounts {
		responses = append(responses, acc.ToAccountResponse())
	}
	return responses
}
