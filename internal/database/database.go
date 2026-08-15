package database

import (
	"context"
	"log"
	"os"

	"wedding_api/ent"

	_ "github.com/lib/pq"
)

func NewDB() *ent.Client {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost port=5432 user=postgres dbname=wedding password=postgres sslmode=disable"
	}

	client, err := ent.Open(
		"postgres",
		dsn,
	)

	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}

	// Cria tabelas se não existirem
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatal("Failed to connect to database", err)
	}
	if err := attachLegacyPresences(context.Background(), client); err != nil {
		log.Fatal("Failed to migrate existing presences", err)
	}

	return client
}

func attachLegacyPresences(ctx context.Context, client *ent.Client) error {
	allPresences, err := client.ConfirmationPresence.Query().WithFamily().All(ctx)
	if err != nil {
		return err
	}
	orphaned := make([]*ent.ConfirmationPresence, 0)
	for _, presence := range allPresences {
		if presence.Edges.Family == nil {
			orphaned = append(orphaned, presence)
		}
	}
	if len(orphaned) == 0 {
		return nil
	}
	legacyFamily, err := client.Family.Create().SetName("Família migrada").Save(ctx)
	if err != nil {
		return err
	}
	for _, presence := range orphaned {
		if err := client.ConfirmationPresence.UpdateOneID(presence.ID).SetFamilyID(legacyFamily.ID).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}
