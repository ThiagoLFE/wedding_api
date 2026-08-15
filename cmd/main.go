package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"wedding_api/internal/database"
	"wedding_api/internal/handlers"
	"wedding_api/internal/router"
	"wedding_api/internal/service"
)

func main() {
	client := database.NewDB()
	defer client.Close()

	service := service.NewService(client)
	if err := service.EnsureAdmin(context.Background(), os.Getenv("ADMIN_EMAIL"), os.Getenv("ADMIN_PASSWORD")); err != nil {
		log.Fatal("failed to bootstrap admin: ", err)
	}
	handler := handlers.NewHandler(service)
	routes := router.NewRouter(handler)

	log.Println("Servidor iniciado em http://localhost:8080")
	http.ListenAndServe(":8080", routes)
}
