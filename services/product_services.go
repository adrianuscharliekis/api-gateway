package services

import (
	"api-gateway/repository"
	"errors"
	"fmt"
)

type ProductService interface {
	IsProductMain(p string, c string) (bool, error)
}

type productService struct {
	productRepository repository.ProductRepository
	tracelogServices  TracelogServices
}

func NewProductService(r repository.ProductRepository, t TracelogServices) ProductService {
	return &productService{productRepository: r, tracelogServices: t}
}

func (s productService) IsProductMain(p string, c string) (bool, error) {
	product, err := s.productRepository.GetProduct(p)
	if err != nil {
		s.tracelogServices.Log("IS PRODUCT MAIN", c, p, err.Error())
		return false, err
	}
	if product.Recid.Valid && product.Recid.String != "" {
		s.tracelogServices.Log("IS PRODUCT MAIN", c, p, "Product Tidak Main")
		return false, errors.New("product tidak main")
	}
	s.tracelogServices.Log("IS PRODUCT MAIN", c, p, fmt.Sprintf("product %s main dan aktif", p))
	return true, nil
}
