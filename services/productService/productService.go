package productService

import (
	"net/http"
	"sample/constans"
	"sample/helpers"
	"sample/models"
	"sample/services"
	"strconv"

	"github.com/labstack/echo"
)

type productService struct {
	Service services.UsecaseService
}

// NewProductService
func NewProductService(service services.UsecaseService) productService {
	return productService{
		Service: service,
	}
}

// AddProduct
func (svc productService) AddProduct(ctx echo.Context) error {
	var result models.Response

	// Validasi dan bind data ke model
	request := new(models.RequestAddProduct)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Setter value ke Model Product
	product := models.Product{
		ProductCode: request.ProductCode,
		ProductName: request.ProductName,
	}

	// Validasi data apakah ada kode produk yg sudah terdaftar pada sistem
	_, exists := svc.Service.ProductRepo.IsProductExistsByIndex(product.ProductCode)
	if exists {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, "Product already exists", nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Proses Insert data produk
	id, err := svc.Service.ProductRepo.AddProduct(product)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	// Proses insert data ke cache
	product.ID = id
	if err := svc.Service.ProductMongoRepo.AddProductMongo(product); err != nil {
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, constans.EMPTY_VALUE, id)

	return ctx.JSON(http.StatusOK, result)
}

// UpdateProduct
func (svc productService) UpdateProduct(ctx echo.Context) error {
	var result models.Response

	request := new(models.RequestUpdateProduct)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	product := models.Product{
		ID:          request.ID,
		ProductCode: request.ProductCode,
		ProductName: request.ProductName,
	}

	_, err := svc.Service.ProductRepo.FindProductById(product.ID)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	id, err := svc.Service.ProductRepo.UpdateProduct(product)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, err.Error(), id)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, constans.EMPTY_VALUE, id)

	return ctx.JSON(http.StatusOK, result)
}

// RemoveProduct
func (svc productService) RemoveProduct(ctx echo.Context) error {
	var result models.Response

	ID, _ := strconv.Atoi(ctx.Param("id"))

	product := models.Product{
		ID: ID,
	}

	_, err := svc.Service.ProductRepo.FindProductById(product.ID)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	err = svc.Service.ProductRepo.RemoveProduct(product.ID)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, constans.EMPTY_VALUE, ID)

	return ctx.JSON(http.StatusOK, result)
}
