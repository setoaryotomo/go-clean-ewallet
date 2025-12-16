package models

import (
	"sample/constans"
	"time"
)

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

// Request model untuk Transaction History List V2
type RequestTransactionHistoryList struct {
	AccountNumber string `json:"account_number"`
	StartDate     string `json:"start_date"` // Format: 2006-01-02
	EndDate       string `json:"end_date"`   // Format: 2006-01-02
	SearchValue   string `json:"search_value"`
	Draw          int    `json:"draw"`
	AscDesc       string `json:"asc_desc"`          // ASC atau DESC
	ColumnOrder   string `json:"column_order_name"` // Nama kolom untuk sorting
	PageNumber    int    `json:"page_number"`
	PageSize      int    `json:"page_size" validate:"required"`
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

type DepositResponse struct {
	AccountNumber   string  `json:"account_number"`
	AccountName     string  `json:"account_name"`
	BalanceBefore   float64 `json:"balance_before"`
	Amount          float64 `json:"amount"`
	BalanceAfter    float64 `json:"balance_after"`
	TransactionDate string  `json:"transaction_date"`
}

type WithdrawResponse struct {
	AccountNumber   string  `json:"account_number"`
	AccountName     string  `json:"account_name"`
	BalanceBefore   float64 `json:"balance_before"`
	Amount          float64 `json:"amount"`
	BalanceAfter    float64 `json:"balance_after"`
	TransactionDate string  `json:"transaction_date"`
}

type TransferResponse struct {
	FromAccountNumber string    `json:"source_number"`
	ToAccountNumber   string    `json:"beneficiary_number"`
	Amount            float64   `json:"amount"`
	FromBalanceAfter  float64   `json:"from_balance_after"`
	FromBalanceBefore float64   `json:"from_balance_before"`
	ToBalanceAfter    float64   `json:"to_balance_after"`
	ToBalanceBefore   float64   `json:"to_balance_before"`
	TransactionDate   time.Time `json:"transaction_date"`
}

type TransactionListResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	TotalRecords int                   `json:"total_records"`
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
	// TransactionTime time.Time `json:"transaction_time"`
	TransactionTime string `json:"transaction_time"`
	// CreatedAt       time.Time `json:"created_at"`
	CreatedAt string `json:"created_at"`
}

type TransactionSimpleResponse struct {
	ID              int    `json:"id"`
	AccountNumber   string `json:"account_number"`
	TransactionType string `json:"transaction_type"`
	// TransactionTypeDesc string    `json:"transaction_type_desc"` // Debit (Keluar) / Credit (Masuk)
	Amount float64 `json:"amount"`
	// TransactionTime time.Time `json:"transaction_time"`
	TransactionTime string `json:"transaction_time"`
}

type TransactionHistorySimpleResponse struct {
	Transactions []TransactionSimpleResponse `json:"transactions"`
	Pagination   PaginationMeta              `json:"pagination"`
}

// Response model untuk Transaction History List V2
type ResponseTransactionHistoryListV2 struct {
	ID            int    `json:"id"`
	AccountID     int    `json:"account_id"`
	AccountNumber string `json:"account_number"`
	AccountName   string `json:"account_name"`
	// SourceNumber      string    `json:"source_number,omitempty"`
	// BeneficiaryNumber string    `json:"beneficiary_number,omitempty"`
	TransactionType string  `json:"transaction_type"` // D atau C
	Amount          float64 `json:"amount"`
	TransactionTime string  `json:"transaction_time"` // Format: YYYY-MM-DD HH:MM:SS
	CreatedAt       string  `json:"created_at"`       // Format: YYYY-MM-DD HH:MM:SS
}

// Response wrapper untuk Transaction History V2
type ResponseTransactionHistoryV2 struct {
	ReferenceNo     string                             `json:"referenceNo"`
	RecordsFiltered int                                `json:"recordsFiltered"`
	RecordsTotal    int                                `json:"recordsTotal"`
	Value           []ResponseTransactionHistoryListV2 `json:"value"`
}

// Result model untuk count dan summaries
type ResultDataTableTransactionCountAndSummaries struct {
	Count          int64   `json:"count"`
	SumariesDebit  float64 `json:"sumaries_debit"`
	SumariesCredit float64 `json:"sumaries_credit"`
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
		// TransactionTime: t.TransactionTime,
		TransactionTime: t.TransactionTime.Format(constans.LAYOUT_TIMESTAMP),
		// CreatedAt:       t.CreatedAt,
		CreatedAt: t.CreatedAt.Format(constans.LAYOUT_TIMESTAMP),
	}
}

func (t *Transaction) ToSimpleResponse() TransactionSimpleResponse {
	return TransactionSimpleResponse{
		ID:              t.ID,
		AccountNumber:   t.AccountNumber,
		TransactionType: t.TransactionType,
		// TransactionTypeDesc: t.GetTypeDescription(),
		Amount:          t.Amount,
		TransactionTime: t.TransactionTime.Format(constans.LAYOUT_TIMESTAMP),
	}
}
