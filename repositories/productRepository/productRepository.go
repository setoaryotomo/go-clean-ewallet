package productRepository

import (
	"database/sql"
	"errors"
	"log"
	"sample/models"
	"sample/repositories"
)

var defineColumn = `id, product_code, product_name`

type productRepository struct {
	RepoDB repositories.Repository
}

// NewProductRepository
func NewProductRepository(repoDB repositories.Repository) productRepository {
	return productRepository{
		RepoDB: repoDB,
	}
}

// FindProductById mencari id primary key
func (ctx productRepository) FindProductById(id int) (models.Product, error) {
	var product models.Product

	var query = `SELECT ` + defineColumn + ` FROM m_product WHERE id = $1`

	rows, err := ctx.RepoDB.DB.Query(query, id)
	if err != nil {
		return product, err
	}
	defer rows.Close()

	data, err := productDto(rows)
	if len(data) == 0 {
		return product, errors.New("Product not found")

	}

	return data[0], nil
}

// FindProductByIndex mencari unique index
func (ctx productRepository) FindProductByIndex(code string) (models.Product, error) {
	var product models.Product

	var query = `SELECT ` + defineColumn + ` FROM m_product WHERE product_code = $1`

	rows, err := ctx.RepoDB.DB.Query(query, code)
	if err != nil {
		return product, err
	}
	defer rows.Close()

	data, err := productDto(rows)
	if len(data) == 0 {
		return product, errors.New("Product not found")

	}

	return data[0], nil
}

// IsProductExistsByIndex
func (ctx productRepository) IsProductExistsByIndex(code string) (models.Product, bool) {
	var product models.Product

	var query = `SELECT ` + defineColumn + ` FROM m_product WHERE product_code = $1`

	rows, err := ctx.RepoDB.DB.Query(query, code)
	if err != nil {
		log.Printf("Error1: %v\n", err)
		return product, false

	}

	data, err := productDto(rows)
	if len(data) == 0 {
		log.Printf("Error2: %v\n", err)
		return product, false

	}

	return data[0], true
}

// AddProduct
func (ctx productRepository) AddProduct(product models.Product) (int, error) {
	var ID int

	query := `INSERT INTO m_product (
				product_code, product_name
		) VALUES (
				$1, $2
		) RETURNING id`

	err := ctx.RepoDB.DB.QueryRow(query, product.ProductCode, product.ProductName).Scan(&ID)
	if err != nil {
		return ID, err
	}

	return ID, nil
}

// UpdateProduct
func (ctx productRepository) UpdateProduct(product models.Product) (int, error) {
	var ID int

	var strQuery = `UPDATE m_product SET product_name = $2 WHERE id = $1 RETURNING id`

	err := ctx.RepoDB.DB.QueryRow(strQuery, product.ID, product.ProductName).Scan(&ID)
	if err != nil {
		return 0, err
	}

	return ID, nil
}

// RemoveProduct
func (ctx productRepository) RemoveProduct(id int) error {

	_, err := ctx.RepoDB.DB.Exec("DELETE FROM m_product WHERE id = $1", id)

	if err != nil {
		return err
	}

	return err
}

// GetProductList
func (ctx productRepository) GetProductList() ([]models.Product, error) {
	var query = `SELECT ` + defineColumn + ` FROM m_product`

	rows, err := ctx.RepoDB.DB.Query(query)
	if err != nil {
		return nil, err
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return productDto(rows)
}

// productDto
func productDto(rows *sql.Rows) ([]models.Product, error) {
	var result []models.Product

	for rows.Next() {
		var val models.Product
		err := rows.Scan(&val.ID, &val.ProductCode, &val.ProductName)
		if err != nil {
			log.Printf("Error: %v\n", err)
			return result, err
		}
		result = append(result, val)
	}

	return result, nil
}
