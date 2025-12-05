package models

type RequestAddProduct struct {
	ProductCode	string	`json:"productCode" validate:"required"`
	ProductName	string	`json:"productName" validate:"required"`
}

type RequestUpdateProduct struct {
	ID			int		`json:"id" validate:"required"`
	ProductCode	string	`json:"productCode" validate:"required"`
	ProductName	string	`json:"productName" validate:"required"`
}

type Product struct {
	ID			int	   `json:"id"`
	ProductCode string `json:"productCode"`
	ProductName string `json:"productName"`
}

type ProductStyle struct {
	ID			int		`json:"id"`
	ProductCode	string	`json:"productCode"`
	StyleItem	string	`json:"styleItem"`
}