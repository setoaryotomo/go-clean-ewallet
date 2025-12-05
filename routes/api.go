package routes

import (
	"net/http"
	"sample/config"
	"sample/services"
	"sample/services/accountService"
	"sample/services/pedagangKiosGradingService"
	"sample/services/pedagangKiosPoinService.go"
	"sample/services/productService"
	"sample/services/warehouseService"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// RoutesApi
func RoutesApi(e echo.Echo, usecaseSvc services.UsecaseService) {

	public := e.Group("/public")

	// ==================== PRODUCT ROUTES ====================
	productSvc := productService.NewProductService(usecaseSvc)
	productGroup := public.Group("/product")
	productGroup.POST("/add", productSvc.AddProduct)
	productGroup.POST("/update", productSvc.UpdateProduct)

	productMongoSvc := productService.NewProductMongoService(usecaseSvc)
	productMongoGroup := public.Group("/product/mongo")
	productMongoGroup.POST("/add", productMongoSvc.AddProductMongo)

	// ==================== WAREHOUSE ROUTES ====================
	warehouseSvc := warehouseService.NewWarehouseService(usecaseSvc)
	warehouseGroup := public.Group("/warehouse")
	warehouseGroup.POST("/add", warehouseSvc.AddWarehouse)

	// ==================== PEDAGANG KIOS POIN ROUTES ====================
	pedagangKiosPoinSvc := pedagangKiosPoinService.NewPedagangKiosPoinService(usecaseSvc)
	pedagangKiosPoinGroup := public.Group("/pedagangkiospoin")
	pedagangKiosPoinGroup.GET("/daily", pedagangKiosPoinSvc.AddPedagangKiosPoinDaily)
	pedagangKiosPoinGroup.GET("/inisiasi", pedagangKiosPoinSvc.AddPedagangKiosPoinInisiasi)

	// ==================== PEDAGANG KIOS GRADING ROUTES ====================
	pedagangKiosGradingSvc := pedagangKiosGradingService.NewPedagangKiosGradingService(usecaseSvc)
	pedagangKiosGradingGroup := public.Group("/pedagangkiosgrading")
	pedagangKiosGradingGroup.GET("/inisiasi", pedagangKiosGradingSvc.AddPedagangKiosGradingWeekly)

	// ==================== ACCOUNT ROUTES ====================
	accountSvc := accountService.NewAccountService(usecaseSvc)
	accountGroup := public.Group("/account")

	// Account Management
	accountGroup.POST("/create", accountSvc.CreateAccount)              // Buat akun baru
	accountGroup.POST("/list", accountSvc.GetAccountList)               // List semua akun (POST)
	accountGroup.POST("/get", accountSvc.GetAccountByID)                // Get akun by ID (POST dengan body)
	accountGroup.GET("/:account_number", accountSvc.GetAccountByNumber) // Detail akun by nomor (GET param)
	accountGroup.POST("/update", accountSvc.UpdateAccount)              // Update data akun
	accountGroup.POST("/delete", accountSvc.DeleteAccount)              // Hapus akun (POST dengan body)

	// Balance Operations
	accountGroup.POST("/balance/check", accountSvc.CheckBalance) // Cek saldo (perlu PIN)
	accountGroup.POST("/deposit", accountSvc.Deposit)            // Setor tunai
	accountGroup.POST("/withdraw", accountSvc.Withdraw)          // Tarik tunai (perlu PIN)

	// Transfer & PIN
	accountGroup.POST("/transfer", accountSvc.Transfer)    // Transfer antar akun (perlu PIN)
	accountGroup.POST("/pin/update", accountSvc.UpdatePIN) // Update PIN

	// ==================== PRIVATE ROUTES (JWT Protected) ====================
	private := e.Group("/private")
	private.Use(middleware.JWT([]byte(config.GetEnv("JWT_KEY"))))
	private.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowCredentials: true,
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

}
