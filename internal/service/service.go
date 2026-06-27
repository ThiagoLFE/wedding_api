package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"wedding_api/ent"
	"wedding_api/ent/confirmationpresence"
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
	if p.Value <= 0.00 {
		return errors.New("product value must be greater than 0")
	}

	return nil
}

type ConfirmationPresenceRequest struct {
	Fullname    string `json:"fullname"`
	PhotoBase64 string `json:"photo_base64"`
	IsConfirmed bool   `json:"is_confirmed"`
}

type UpdateConfirmationPresence struct {
	ID int `json:"id"`
	ConfirmationPresenceRequest
}

func (p *ConfirmationPresenceRequest) Validate() error {
	if len(strings.TrimSpace(p.Fullname)) == 0 {
		return errors.New("fullname is required")
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
	products, err := s.client.Product.Query().Order(product.ByID()).All(ctx)
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

func (s *Service) AddConfirmationPresence(ctx context.Context, presence ConfirmationPresenceRequest) (*ent.ConfirmationPresence, error) {
	if err := presence.Validate(); err != nil {
		return nil, err
	}

	createdPresence, err := s.client.ConfirmationPresence.Create().
		SetFullname(strings.TrimSpace(presence.Fullname)).
		SetPhotoBase64(presence.PhotoBase64).
		SetIsConfirmed(presence.IsConfirmed).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return createdPresence, nil
}

func (s *Service) ConfirmationPresencesList(ctx context.Context) ([]*ent.ConfirmationPresence, error) {
	presences, err := s.client.ConfirmationPresence.Query().Order(confirmationpresence.ByID()).All(ctx)
	if err != nil {
		return nil, err
	}

	return presences, nil
}

func (s *Service) GetConfirmationPresence(ctx context.Context, id int) (*ent.ConfirmationPresence, error) {
	presence, err := s.client.ConfirmationPresence.Query().
		Where(confirmationpresence.IDEQ(id)).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	return presence, nil
}

func (s *Service) UpdateConfirmationPresence(ctx context.Context, newPresence UpdateConfirmationPresence) (*ent.ConfirmationPresence, error) {
	if err := newPresence.ConfirmationPresenceRequest.Validate(); err != nil {
		return nil, err
	}

	updatedPresence, err := s.client.ConfirmationPresence.UpdateOneID(newPresence.ID).
		SetFullname(strings.TrimSpace(newPresence.Fullname)).
		SetPhotoBase64(newPresence.PhotoBase64).
		SetIsConfirmed(newPresence.IsConfirmed).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update confirmation presence: %w", err)
	}

	return updatedPresence, nil
}

func (s *Service) ConfirmPresence(ctx context.Context, id int) (*ent.ConfirmationPresence, error) {
	presence, err := s.client.ConfirmationPresence.UpdateOneID(id).
		SetIsConfirmed(true).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to confirm presence: %w", err)
	}

	return presence, nil
}

func (s *Service) CancelPresence(ctx context.Context, id int) (*ent.ConfirmationPresence, error) {
	presence, err := s.client.ConfirmationPresence.UpdateOneID(id).
		SetIsConfirmed(false).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel presence: %w", err)
	}

	return presence, nil
}

func (s *Service) DeleteConfirmationPresence(ctx context.Context, id int) error {
	if err := s.client.ConfirmationPresence.DeleteOneID(id).Exec(ctx); err != nil {
		return err
	}
	return nil
}
