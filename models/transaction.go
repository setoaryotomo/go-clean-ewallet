package models

import "time"

// Transaction model
type Transaction struct {
	ID                int       `json:"id"`
	AccountID         int       `json:"account_id"`
	AccountNumber     string    `json:"account_number"`
	AccountName       string    `json:"account_name"`
	SourceNumber      string    `json:"source_number,omitempty"`
	BeneficiaryNumber string    `json:"beneficiary_number,omitempty"`
	TransactionType   string    `json:"transaction_type"` // 'D' for Debit, 'C' for Credit
	Amount            float64   `json:"amount"`
	TransactionTime   time.Time `json:"transaction_time"`
	CreatedAt         time.Time `json:"created_at"`
}

// Request Models

type RequestTransfer struct {
	FromAccountNumber string  `json:"source_number" validate:"required"`
	ToAccountNumber   string  `json:"beneficiary_number" validate:"required"`
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

type TransactionResponse struct {
	ID                  int       `json:"id"`
	AccountNumber       string    `json:"account_number"`
	AccountName         string    `json:"account_name"`
	SourceNumber        string    `json:"source_number,omitempty"`
	BeneficiaryNumber   string    `json:"beneficiary_number,omitempty"`
	TransactionType     string    `json:"transaction_type"`
	TransactionTypeDesc string    `json:"transaction_type_desc"` // Debit (Keluar) / Credit (Masuk)
	Amount              float64   `json:"amount"`
	Description         string    `json:"description"` // Deskripsi transaksi
	TransactionTime     time.Time `json:"transaction_time"`
}

type TransferResponse struct {
	// TransactionID     string    `json:"transaction_id"`
	FromAccountNumber string    `json:"source_number"`      //from_account_number
	ToAccountNumber   string    `json:"beneficiary_number"` //to_account_number
	Amount            float64   `json:"amount"`
	TransactionDate   time.Time `json:"transaction_date"`
}

// Helper function untuk membuat deskripsi transaksi
func (t *Transaction) GetDescription() string {
	switch t.TransactionType {
	case "D": // Debit (Keluar)
		if t.BeneficiaryNumber != "" {
			return "Transfer ke " + t.BeneficiaryNumber
		}
		return "Penarikan Tunai"
	case "C": // Credit (Masuk)
		if t.SourceNumber != "" {
			return "Transfer dari " + t.SourceNumber
		}
		return "Setoran Tunai"
	default:
		return "Transaksi"
	}
}

// Helper function untuk type description
func (t *Transaction) GetTypeDescription() string {
	switch t.TransactionType {
	case "D":
		return "Debit (Keluar)"
	case "C":
		return "Credit (Masuk)"
	default:
		return "Unknown"
	}
}

// Convert Transaction to TransactionResponse
func (t *Transaction) ToResponse() TransactionResponse {
	return TransactionResponse{
		ID:                  t.ID,
		AccountNumber:       t.AccountNumber,
		AccountName:         t.AccountName,
		SourceNumber:        t.SourceNumber,
		BeneficiaryNumber:   t.BeneficiaryNumber,
		TransactionType:     t.TransactionType,
		TransactionTypeDesc: t.GetTypeDescription(),
		Amount:              t.Amount,
		Description:         t.GetDescription(),
		TransactionTime:     t.TransactionTime,
	}
}
