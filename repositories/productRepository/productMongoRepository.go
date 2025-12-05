package productRepository

import (
	"sample/constans"
	"sample/models"
	"sample/repositories"
)

type productMongoRepository struct {
	RepoDB repositories.Repository
}

func NewProductMongoRepository(repoDB repositories.Repository) productMongoRepository {
	return productMongoRepository{
		RepoDB: repoDB,
	}
}

func (ctx productMongoRepository) AddProductMongo(product models.Product) error {

	_, err := ctx.RepoDB.MongoDB.Collection(constans.PRODUCT_COLLECTION).InsertOne(ctx.RepoDB.Context, product)
	return err
}
