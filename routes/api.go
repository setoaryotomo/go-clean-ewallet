package routes

import (
	"net/http"
	"sample/config"
	"sample/services"
	"sample/services/accountService"
	"sample/services/pedagangKiosGradingService"
	"sample/services/pedagangKiosPoinService.go"
	"sample/services/productService"
	"sample/services/transactionService"
	"sample/services/warehouseService"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// RoutesApi
func RoutesApi(e echo.Echo, usecaseSvc services.UsecaseService) {

	public := e.Group("/public")

	productSvc := productService.NewProductService(usecaseSvc)
	productGroup := public.Group("/product")
	productGroup.POST("/add", productSvc.AddProduct)
	productGroup.POST("/update", productSvc.UpdateProduct)

	productMongoSvc := productService.NewProductMongoService(usecaseSvc)
	productMongoGroup := public.Group("/product/mongo")
	productMongoGroup.POST("/add", productMongoSvc.AddProductMongo)

	warehouseSvc := warehouseService.NewWarehouseService(usecaseSvc)
	warehouseGroup := public.Group("/warehouse")
	warehouseGroup.POST("/add", warehouseSvc.AddWarehouse)

	pedagangKiosPoinSvc := pedagangKiosPoinService.NewPedagangKiosPoinService(usecaseSvc)
	pedagangKiosPoinGroup := public.Group("/pedagangkiospoin")
	pedagangKiosPoinGroup.GET("/daily", pedagangKiosPoinSvc.AddPedagangKiosPoinDaily)
	pedagangKiosPoinGroup.GET("/inisiasi", pedagangKiosPoinSvc.AddPedagangKiosPoinInisiasi)

	pedagangKiosGradingSvc := pedagangKiosGradingService.NewPedagangKiosGradingService(usecaseSvc)
	pedagangKiosGradingGroup := public.Group("/pedagangkiosgrading")
	pedagangKiosGradingGroup.GET("/inisiasi", pedagangKiosGradingSvc.AddPedagangKiosGradingWeekly)

	accountSvc := accountService.NewAccountService(usecaseSvc)
	accountGroup := public.Group("/account")

	// Account
	accountGroup.POST("/create", accountSvc.CreateAccount) // Buat akun baru
	accountGroup.POST("/list", accountSvc.GetAccountList)  // List semua akun
	accountGroup.POST("/get", accountSvc.GetAccountByID)   // Get akun by ID
	accountGroup.POST("/update", accountSvc.UpdateAccount) // Update data akun
	accountGroup.POST("/delete", accountSvc.DeleteAccount) // Hapus akun

	// Transaction
	transactionSvc := transactionService.NewTransactionService(usecaseSvc)
	transactionGroup := public.Group("/transaction")

	transactionGroup.POST("/deposit", transactionSvc.Deposit)   // Setor tunai
	transactionGroup.POST("/withdraw", transactionSvc.Withdraw) // Tarik tunai
	transactionGroup.POST("/transfer", transactionSvc.Transfer) // Transfer antar akun

	private := e.Group("/private")
	private.Use(middleware.JWT([]byte(config.GetEnv("JWT_KEY"))))
	private.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowCredentials: true,
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	//
	// privateAccountGroup := private.Group("/account")
	// privateAccountGroup.GET("/admin/list", accountSvc.GetAccountList)
}
