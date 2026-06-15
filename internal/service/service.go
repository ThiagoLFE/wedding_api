package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"wedding_api/ent"
	"wedding_api/ent/product"
)

type Service struct {
	client *ent.Client
}

func NewService(client *ent.Client) *Service {
	return &Service{
		client: client,
	}
}

type ProductRequest struct {
	Title      string  `json:"title"`
	ReservedBy string  `json:"reserved_by"`
	Image      string  `json:"image"`
	Value      float64 `json:"value"`
}

type UpdateProduct struct {
	ID int `json:"id"`
	ProductRequest
}

func (p *ProductRequest) Validate() error {
	if len(strings.TrimSpace(p.Title)) == 0 {
		return errors.New("the Product must have a title")
	}
	if p.ReservedBy != "" {
		return errors.New("you can't create a present already reserved")
	}
	if p.Value <= 0.00 {
		return errors.New("product value must be greater than 0")
	}

	return nil
}

func (s *Service) AddProduct(ctx context.Context, product ProductRequest) (*ent.Product, error) {
	if err := product.Validate(); err != nil {
		return nil, err
	}
	prod, err := s.client.Product.Create().
		SetTitle(product.Title).
		SetValue(product.Value).
		SetReservedBy("").
		SetImage(product.Image).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return prod, err
}

func (s *Service) ProductsList(ctx context.Context) ([]*ent.Product, error) {
	products, err := s.client.Product.Query().All(ctx)
	if err != nil {
		return nil, err
	}

	return products, nil
}

func (s *Service) GetProduct(ctx context.Context, id int) (*ent.Product, error) {
	products, err := s.client.Product.Query().
		Where(product.IDEQ(id)).
		Only(ctx)

	if err != nil {
		return nil, err
	}

	return products, nil
}

func (s *Service) UpdateProduct(ctx context.Context, newProduct UpdateProduct) (*ent.Product, error) {
	if err := newProduct.ProductRequest.Validate(); err != nil {
		return nil, err
	}

	updProd, err := s.client.Product.UpdateOneID(newProduct.ID).
		SetTitle(newProduct.Title).
		SetValue(newProduct.Value).
		SetReservedBy(newProduct.ReservedBy).
		SetImage(newProduct.Image).
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return updProd, nil
}

func (s *Service) DeleteProduct(ctx context.Context, id int) error {
	if err := s.client.Product.DeleteOneID(id).Exec(ctx); err != nil {
		return err
	}
	return nil
}
