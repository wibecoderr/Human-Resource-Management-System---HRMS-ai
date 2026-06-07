package main

import (
	"hrms/database"
	"hrms/router"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment")
	}

	database.Connect()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := router.New()

	log.Printf("Server running on http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
