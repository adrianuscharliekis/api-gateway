package repository

import (
	"api-gateway/model"
	"database/sql"
)

type ProductRepository interface {
	GetProduct(product string) (*model.Product, error)
}

type productRepository struct {
	db *sql.DB
}

func NewProductRepository(d *sql.DB) ProductRepository {
	return &productRepository{db: d}
}

func (r *productRepository) GetProduct(p string) (*model.Product, error) {
	stmt, err := r.db.Prepare(`
		SELECT productName, recid, path FROM master_product WHERE productName = ?
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	product := &model.Product{}

	err = stmt.QueryRow(p).Scan(&product.ProductName, &product.Recid, &product.Path)
	if err != nil {
		return nil, err
	}
	return product, nil
}
