package service

import "testing"

func TestProductRequestValidate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		req := ProductRequest{Title: "Mesa", Value: 10}
		if err := req.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty title", func(t *testing.T) {
		req := ProductRequest{Title: "   ", Value: 10}
		if err := req.Validate(); err == nil {
			t.Fatal("expected validation error")
		}
	})

	t.Run("non-positive value", func(t *testing.T) {
		req := ProductRequest{Title: "Mesa", Value: 0}
		if err := req.Validate(); err == nil {
			t.Fatal("expected validation error")
		}
	})
}

func TestConfirmationPresenceRequestValidate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		req := ConfirmationPresenceRequest{Fullname: "Joao"}
		if err := req.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty fullname", func(t *testing.T) {
		req := ConfirmationPresenceRequest{Fullname: "   "}
		if err := req.Validate(); err == nil {
			t.Fatal("expected validation error")
		}
	})
}
