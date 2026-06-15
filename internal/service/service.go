package service

import (
	"context"
	"errors"
	"strings"
	"wedding_api/ent"
)

type Service struct {
	client *ent.Client
}

func NewService(client *ent.Client) *Service {
	return &Service{
		client: client,
	}
}

type CreateProduct struct {
	Title      string  `json:"title"`
	ReservedBy string  `json:"reserved_by"`
	Image      string  `json:"image"`
	Value      float64 `json:"value"`
}

func (p *CreateProduct) Validate() error {
	if len(strings.TrimSpace(p.Title)) == 0 {
		return errors.New("The Product must have a title")
	}
	if p.ReservedBy != "" {
		return errors.New("You can't create a present already reserved")
	}
	if p.Value <= 0.00 {
		return errors.New("The value of product must be greater than 0")
	}

	return nil
}

func (h *Service) AddProduct(ctx context.Context, product CreateProduct) (*ent.Product, error) {
	if err := product.Validate(); err != nil {
		return nil, err
	}
	prod, err := h.client.Product.Create().
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
