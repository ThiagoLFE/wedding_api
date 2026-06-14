package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type Product struct {
	ent.Schema
}

// Fields of the Product.
func (Product) Fields() []ent.Field {
	return []ent.Field{
		field.String("title"),
		field.String("reserved_by").Nillable(),
		field.Text("image").Optional(),
		field.Float("value"),
	}
}

// Edges of the Product.
func (Product) Edges() []ent.Edge {
	return nil
}
