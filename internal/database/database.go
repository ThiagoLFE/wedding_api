package database

import (
	"context"
	"log"

	"wedding_api/ent"

	_ "github.com/lib/pq"
)

func NewDB() *ent.Client {

	client, err := ent.Open(
		"postgres",
		"host=localhost port=5432 user=postgres dbname=wedding password=postgres sslmode=disable",
	)

	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}

	defer client.Close()

	// Cria tabelas se não existirem
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatal("Failed to connect to database", err)
	}

	return client
}
