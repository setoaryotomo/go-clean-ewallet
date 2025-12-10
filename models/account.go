package models

import "time"

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

// Request Models

type RequestCreateAccount struct {
	AccountName    string  `json:"account_name" validate:"required,min=3,max=255"`
	PIN            string  `json:"pin" validate:"required,len=6"`
	InitialDeposit float64 `json:"initial_deposit"`
}

type RequestUpdateAccount struct {
	ID          int    `json:"id" validate:"required"`
	AccountName string `json:"account_name" validate:"required,min=3,max=255"`
}

type RequestDeleteAccount struct {
	ID int `json:"id" validate:"required,min=1"`
}

type RequestUpdateBalance struct {
	AccountNumber string  `json:"account_number" validate:"required"`
	Amount        float64 `json:"amount" validate:"required"`
}

type RequestChangePIN struct {
	AccountNumber string `json:"account_number" validate:"required"`
	OldPIN        string `json:"old_pin" validate:"required,len=6"`
	NewPIN        string `json:"new_pin" validate:"required,len=6"`
}

// UPDATED: Hanya membutuhkan account_number
type RequestForgotPIN struct {
	AccountNumber string `json:"account_number" validate:"required"`
}

// UPDATED: Membutuhkan reset_token, new_pin, dan confirm_new_pin
type RequestResetPIN struct {
	ResetToken    string `json:"reset_token" validate:"required"`
	NewPIN        string `json:"new_pin" validate:"required,len=6"`
	ConfirmNewPIN string `json:"confirm_new_pin" validate:"required,len=6"`
}

// Response Models

type AccountResponse struct {
	ID            int       `json:"id"`
	AccountNumber string    `json:"account_number"`
	Balance       float64   `json:"balance"`
	AccountName   string    `json:"account_name"`
	AccountStatus string    `json:"account_status"`
	CreatedAt     time.Time `json:"created_at"`
}

type BalanceResponse struct {
	AccountNumber string  `json:"account_number"`
	AccountName   string  `json:"account_name"`
	Balance       float64 `json:"balance"`
}

type RequestGetAccountByID struct {
	ID int `json:"id" validate:"required,min=1"`
}

type AccountDetailResponse struct {
	ID            int       `json:"id"`
	AccountNumber string    `json:"account_number"`
	AccountName   string    `json:"account_name"`
	Balance       float64   `json:"balance"`
	AccountStatus string    `json:"account_status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	IsDeleted     bool      `json:"is_deleted,omitempty"`
}

type ChangePINResponse struct {
	AccountNumber string    `json:"account_number"`
	ChangedAt     time.Time `json:"changed_at"`
}

type ForgotPINResponse struct {
	AccountNumber string    `json:"account_number"`
	ResetToken    string    `json:"reset_token"`
	ExpiresAt     time.Time `json:"expires_at"`
}

type ResetPINResponse struct {
	AccountNumber string    `json:"account_number"`
	ResetAt       time.Time `json:"reset_at"`
}
