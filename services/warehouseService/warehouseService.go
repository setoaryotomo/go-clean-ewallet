package warehouseService

import (
	"github.com/labstack/echo"
	"net/http"
	"sample/constans"
	"sample/helpers"
	"sample/models"
	"sample/services"
)

type warehouseService struct {
	Service services.UsecaseService
}

func NewWarehouseService(service services.UsecaseService) warehouseService  {
	return warehouseService{
		Service: service,
	}
}

func (svc warehouseService) AddWarehouse(ctx echo.Context) error {
	var result models.Response

	// Validasi dan bind data ke model
	request := new(models.RequestWarehouse)
	if err := helpers.BindValidateStruct(ctx, request); err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadGateway, result)
	}
	helpers.LOG("Incoming Request Add Warehouse: ", request, false)

	return nil
}
