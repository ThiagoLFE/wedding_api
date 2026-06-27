package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// ConfirmationPresence holds the schema definition for invited people presence confirmation.
type ConfirmationPresence struct {
	ent.Schema
}

// Fields of the ConfirmationPresence.
func (ConfirmationPresence) Fields() []ent.Field {
	return []ent.Field{
		field.String("fullname"),
		field.Text("photo_base64").Optional(),
		field.Bool("is_confirmed").Default(false),
	}
}

// Edges of the ConfirmationPresence.
func (ConfirmationPresence) Edges() []ent.Edge {
	return nil
}
