package main

import (
	"log"
	"net/http"

	"wedding_api/internal/database"
	"wedding_api/internal/handlers"
	"wedding_api/internal/service"
)

func main() {

	client := database.GetDBX()

	svc := service.GetService(client)

	handler := handlers.GetHandler(svc)
	log.Println("Servidor iniciado em :8080")

	http.ListenAndServe(":8080", r)
}
