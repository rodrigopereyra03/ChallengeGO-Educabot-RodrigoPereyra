package main

import (
	"net/http"
	"time"

	"educabot.com/bookshop/cmd/app/handlers"
	"educabot.com/bookshop/cmd/app/router"
	"educabot.com/bookshop/internal/books/providers"
	"educabot.com/bookshop/internal/books/services"
)

func main() {
	httpClient := &http.Client{Timeout: 5 * time.Second}

	bookProvider := providers.NewExternalBooksProvider(httpClient,
		"https://6781684b85151f714b0aa5db.mockapi.io/api/v1/books",
	)

	bookService := services.NewMetricsService(bookProvider)

	bookHandler := handlers.NewMetricsHandler(bookService)

	r := router.NewRouter(bookHandler).Setup()

	r.Run(":8080")
}
