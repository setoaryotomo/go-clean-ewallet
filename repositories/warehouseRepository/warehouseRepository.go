package warehouseRepository

import (
	"sample/models"
	"sample/repositories"
)

type warehouseRepository struct {
	RepoDB repositories.Repository
}

func NewWarehouseRepository(repoDB repositories.Repository) warehouseRepository {
	return warehouseRepository{
		RepoDB: repoDB,
	}
}

func (ctx warehouseRepository) FindWarehouseById(id int) (models.Warehouse, error) {
	var warehouse models.Warehouse

	return warehouse, nil
}