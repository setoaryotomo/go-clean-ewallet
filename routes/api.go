package routes

import (
	"net/http"
	"sample/config"
	"sample/services"
	"sample/services/accountService"
	"sample/services/transactionHistoryService"
	"sample/services/transactionService"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// RoutesApi
func RoutesApi(e echo.Echo, usecaseSvc services.UsecaseService) {

	public := e.Group("/public")

	// ============================================
	// Account Service (PostgreSQL)
	// ============================================
	accountSvc := accountService.NewAccountService(usecaseSvc)
	accountGroup := public.Group("/account")

	// Account Management
	accountGroup.POST("/create", accountSvc.CreateAccount) // Buat akun baru
	accountGroup.POST("/list", accountSvc.GetAccountList)  // List semua akun
	accountGroup.POST("/get", accountSvc.GetAccountByID)   // Get akun by ID
	accountGroup.POST("/update", accountSvc.UpdateAccount) // Update data akun
	accountGroup.POST("/delete", accountSvc.DeleteAccount) // Hapus akun

	accountGroup.POST("/balance-inquiry", accountSvc.GetBalanceInquiry)

	// PIN Management
	accountGroup.POST("/change-pin", accountSvc.ChangePIN) // Ubah PIN
	accountGroup.POST("/forgot-pin", accountSvc.ForgotPIN) // Lupa PIN - Generate reset token
	accountGroup.POST("/reset-pin", accountSvc.ResetPIN)   // Reset PIN dengan token

	// ============================================
	// Transaction Service
	// ============================================
	transactionSvc := transactionService.NewTransactionService(usecaseSvc)
	transactionHistorySvc := transactionHistoryService.NewTransactionHistoryService(usecaseSvc)
	transactionGroup := public.Group("/transaction")

	// Basic Transactions
	transactionGroup.POST("/deposit", transactionSvc.Deposit)   // Setor tunai
	transactionGroup.POST("/withdraw", transactionSvc.Withdraw) // Tarik tunai
	transactionGroup.POST("/transfer", transactionSvc.Transfer) // Transfer antar akun

	// Transaction History
	transactionGroup.POST("/history-v2", transactionHistorySvc.TransactionHistoryListV2) // Riwayat transaksi
	transactionGroup.POST("/history", transactionSvc.GetTransactionHistory)              // Riwayat transaksi
	transactionGroup.POST("/detail", transactionSvc.GetTransactionDetail)                // Detail transaksi

	// ============================================
	// Private Routes (Authenticated)
	// ============================================
	private := e.Group("/private")
	private.Use(middleware.JWT([]byte(config.GetEnv("JWT_KEY"))))
	private.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowCredentials: true,
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Private routes can be added here for admin/authenticated users
	// privateAccountGroup := private.Group("/account")
	// privateAccountGroup.GET("/admin/list", accountSvc.GetAccountList)

}
