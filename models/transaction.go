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
	PIN           string  `json:"pin" validate:"required,len=6"`
}

type RequestWithdraw struct {
	AccountNumber string  `json:"account_number" validate:"required"`
	Amount        float64 `json:"amount" validate:"required,min=10000"`
	PIN           string  `json:"pin" validate:"required,len=6"`
}

type RequestTransactionHistory struct {
	AccountNumber string `json:"account_number,omitempty"`
	StartDate     string `json:"start_date,omitempty" validate:"required"` // Format: 2006-01-02
	EndDate       string `json:"end_date,omitempty" validate:"required"`   // Format: 2006-01-02
	Limit         int    `json:"limit,omitempty" validate:"required"`      // Default 10
	Page          int    `json:"page,omitempty" validate:"required"`       // Default 1
}

type RequestTransactionDetail struct {
	TransactionID int `json:"transaction_id" validate:"required,min=1"`
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
	FromAccountNumber string    `json:"source_number"`
	ToAccountNumber   string    `json:"beneficiary_number"`
	Amount            float64   `json:"amount"`
	TransactionDate   time.Time `json:"transaction_date"`
}

type TransactionHistoryResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	TotalRecords int                   `json:"total_records"`
	Pagination   PaginationMeta        `json:"pagination"`
}

type PaginationMeta struct {
	CurrentPage  int `json:"current_page"`
	PerPage      int `json:"per_page"`
	TotalRecords int `json:"total_records"`
	TotalPages   int `json:"total_pages"`
}

type TransactionDetailResponse struct {
	ID                int    `json:"id"`
	AccountNumber     string `json:"account_number"`
	AccountName       string `json:"account_name"`
	SourceNumber      string `json:"source_number,omitempty"`
	BeneficiaryNumber string `json:"beneficiary_number,omitempty"`
	TransactionType   string `json:"transaction_type"`
	// TransactionTypeDesc string    `json:"transaction_type_desc"`
	Amount float64 `json:"amount"`
	// Description     string    `json:"description"`
	TransactionTime time.Time `json:"transaction_time"`
	CreatedAt       time.Time `json:"created_at"`
}

type TransactionSimpleResponse struct {
	ID              int    `json:"id"`
	AccountNumber   string `json:"account_number"`
	TransactionType string `json:"transaction_type"`
	// TransactionTypeDesc string    `json:"transaction_type_desc"` // Debit (Keluar) / Credit (Masuk)
	Amount          float64   `json:"amount"`
	TransactionTime time.Time `json:"transaction_time"`
}

type TransactionHistorySimpleResponse struct {
	Transactions []TransactionSimpleResponse `json:"transactions"`
	Pagination   PaginationMeta              `json:"pagination"`
}

// Helper function untuk membuat deskripsi transaksi
func (t *Transaction) GetDescription() string {
	switch t.TransactionType {
	case "D": // Debit (Keluar)
		if t.BeneficiaryNumber != "" && t.BeneficiaryNumber != t.AccountNumber {
			return "Transfer ke " + t.BeneficiaryNumber
		}
		return "Penarikan Tunai"
	case "C": // Credit (Masuk)
		if t.SourceNumber != "" && t.SourceNumber != t.AccountNumber {
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

// Convert Transaction to TransactionDetailResponse
func (t *Transaction) ToDetailResponse() TransactionDetailResponse {
	return TransactionDetailResponse{
		ID:                t.ID,
		AccountNumber:     t.AccountNumber,
		AccountName:       t.AccountName,
		SourceNumber:      t.SourceNumber,
		BeneficiaryNumber: t.BeneficiaryNumber,
		TransactionType:   t.TransactionType,
		// TransactionTypeDesc: t.GetTypeDescription(),
		Amount: t.Amount,
		// Description:     t.GetDescription(),
		TransactionTime: t.TransactionTime,
		CreatedAt:       t.CreatedAt,
	}
}

func (t *Transaction) ToSimpleResponse() TransactionSimpleResponse {
	return TransactionSimpleResponse{
		ID:              t.ID,
		AccountNumber:   t.AccountNumber,
		TransactionType: t.TransactionType,
		// TransactionTypeDesc: t.GetTypeDescription(),
		Amount:          t.Amount,
		TransactionTime: t.TransactionTime,
	}
}
