package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	portString := os.Getenv("PORT")

	if portString == "" {
		log.Fatal("PORT is not found in enviroment variable")
	}

	router := chi.NewRouter()

	srv := &http.Server{
		Addr:    ":" + portString,
		Handler: router,
	}
	log.Printf("Server starting at port %s\n", portString)
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal("Error in starting server", err)
	}
}
