package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Family represents a family invited to the wedding.
type Family struct {
	ent.Schema
}

func (Family) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
	}
}

func (Family) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("presences", ConfirmationPresence.Type),
		edge.To("access_tokens", FamilyAccessToken.Type),
		edge.To("sessions", Session.Type),
	}
}
