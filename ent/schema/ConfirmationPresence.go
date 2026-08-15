package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
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
	return []ent.Edge{
		// Kept nullable for the first migration so existing phase-1 rows can be
		// assigned to a migration family before new rows are created.
		edge.From("family", Family.Type).Ref("presences").Unique(),
	}
}
