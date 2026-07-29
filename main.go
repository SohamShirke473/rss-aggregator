package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"github.com/SohamShirke473/rss-aggregator/internal/database"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB *database.Queries
}

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Warning: .env file not found or failed to load: %v", err)
	}

	portString := os.Getenv("PORT")

	if portString == "" {
		log.Fatal("PORT is not found in enviroment variable")
	}
	dbURL := os.Getenv("DATABASE_URL")

	if dbURL == "" {
		log.Fatal("DATABASE_URL is not found in enviroment variable")
	}

	conn, err := sql.Open("postgres", dbURL)

	if err != nil {
		log.Fatal("Error in opening database", err)
	}

	apiCfg := apiConfig{
		DB: database.New(conn),
	}

	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://*", "https://*"},
		AllowedMethods:   []string{"GET", "PUT", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	v1Router := chi.NewRouter()
	v1Router.Get("/ready", handlerReadiness)
	v1Router.Get("/err", handlerErr)
	v1Router.Post("/users", apiCfg.handlerCreateUser)
	v1Router.Get("/users", apiCfg.middlewareAuth(apiCfg.handlerGetUser))

	v1Router.Post("/feeds", apiCfg.middlewareAuth(apiCfg.handlerCreateFeed))
	v1Router.Get("/feeds", apiCfg.handlerGetFeeds)
	router.Mount("/v1", v1Router)

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
