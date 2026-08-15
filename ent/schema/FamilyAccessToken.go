package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// FamilyAccessToken is a revocable bearer link used to create family sessions.
// A nil expiration means the link does not expire automatically.
type FamilyAccessToken struct {
	ent.Schema
}

func (FamilyAccessToken) Fields() []ent.Field {
	return []ent.Field{
		field.String("token_hash").Unique(),
		field.Time("expires_at").Optional().Nillable(),
		field.Time("revoked_at").Optional().Nillable(),
		field.Time("last_used_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now),
	}
}

func (FamilyAccessToken) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("family", Family.Type).Ref("access_tokens").Unique().Required(),
	}
}
