package models

type RequestWarehouse struct {
	ID int `json:"id"`
	WarehouseCode string `json:"warehouseCode" validate:"required"`
	WarehouseName string `json:"warehouseName" validate:"required"`
}

type Warehouse struct {
	ID int `json:"id"`
	WarehouseCode string `json:"warehouseCode"`
	WarehouseName string `json:"warehouseName"`
}