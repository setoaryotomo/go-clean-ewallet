package productService

import (
	"github.com/labstack/echo"
	"net/http"
	"sample/constans"
	"sample/helpers"
	"sample/models"
	"sample/services"
)

type productMongoService struct {
	Service services.UsecaseService
}

func NewProductMongoService(service services.UsecaseService) productMongoService {
	return productMongoService{
		Service: service,
	}
}

func (svc productMongoService) AddProductMongo(ctx echo.Context) error {
	var result models.Response

	request := new(models.RequestAddProduct)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
	}
	helpers.LOG("Incoming Request AddProductMongo", request, false)

	product := models.Product{
		ProductCode: request.ProductCode,
		ProductName: request.ProductName,
	}

	err := svc.Service.ProductMongoRepo.AddProductMongo(product)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	result = helpers.ResponseJSON(false, constans.SUCCESS_CODE, constans.EMPTY_VALUE, nil)

	return ctx.JSON(http.StatusOK, result)
}