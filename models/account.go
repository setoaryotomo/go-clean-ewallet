package models

import "time"

// Account model
type Account struct {
	ID            int       `json:"id"`
	AccountNumber string    `json:"account_number"`
	Balance       float64   `json:"balance"`
	PIN           string    `json:"pin,omitempty"` // omitempty agar tidak muncul di response
	AccountName   string    `json:"account_name"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Request Models

type RequestCreateAccount struct {
	// AccountNumber  string  `json:"account_number" validate:"required,min=10,max=20"`
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

type RequestUpdatePIN struct {
	AccountNumber string `json:"account_number" validate:"required"`
	OldPIN        string `json:"old_pin" validate:"required,len=6"`
	NewPIN        string `json:"new_pin" validate:"required,len=6"`
}

type RequestTransfer struct {
	FromAccountNumber string  `json:"from_account_number" validate:"required"`
	ToAccountNumber   string  `json:"to_account_number" validate:"required"`
	Amount            float64 `json:"amount" validate:"required,min=10000"`
	PIN               string  `json:"pin" validate:"required,len=6"`
}

type RequestCheckBalance struct {
	AccountNumber string `json:"account_number" validate:"required"`
	PIN           string `json:"pin" validate:"required,len=6"`
}

type RequestDeposit struct {
	AccountNumber string  `json:"account_number" validate:"required"`
	Amount        float64 `json:"amount" validate:"required,min=10000"`
}

type RequestWithdraw struct {
	AccountNumber string  `json:"account_number" validate:"required"`
	Amount        float64 `json:"amount" validate:"required,min=10000"`
	PIN           string  `json:"pin" validate:"required,len=6"`
}

// Response Models

type AccountResponse struct {
	ID            int       `json:"id"`
	AccountNumber string    `json:"account_number"`
	Balance       float64   `json:"balance"`
	AccountName   string    `json:"account_name"`
	CreatedAt     time.Time `json:"created_at"`
}

type BalanceResponse struct {
	AccountNumber string  `json:"account_number"`
	AccountName   string  `json:"account_name"`
	Balance       float64 `json:"balance"`
}

type TransferResponse struct {
	TransactionID     string    `json:"transaction_id"`
	FromAccountNumber string    `json:"from_account_number"`
	ToAccountNumber   string    `json:"to_account_number"`
	Amount            float64   `json:"amount"`
	TransactionDate   time.Time `json:"transaction_date"`
}

// RequestGetAccountByID untuk POST method
type RequestGetAccountByID struct {
	ID int `json:"id" validate:"required,min=1"`
}

// AccountDetailResponse untuk response detail akun
type AccountDetailResponse struct {
	ID            int       `json:"id"`
	AccountNumber string    `json:"account_number"`
	AccountName   string    `json:"account_name"`
	Balance       float64   `json:"balance"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	IsDeleted     bool      `json:"is_deleted,omitempty"`
}
