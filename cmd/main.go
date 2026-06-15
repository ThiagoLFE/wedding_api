package main

import (
	"log"
	"net/http"

	"wedding_api/internal/database"
	"wedding_api/internal/handlers"
	"wedding_api/internal/router"
	"wedding_api/internal/service"
)

func main() {
	client := database.NewDB()
	defer client.Close()

	service := service.NewService(client)
	handler := handlers.NewHandler(service)
	routes := router.NewRouter(handler)

	log.Println("Servidor iniciado em http://localhost:8080/")
	http.ListenAndServe(":8080", routes)
}
