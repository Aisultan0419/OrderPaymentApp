// @title           Payment Service API
// @version         1.0
// @description     Processes payments and validates transaction limits. Amounts above $1000 (100000 cents) are declined.
// @host            localhost:8081
// @BasePath        /

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "payment-service/docs" // swagger generated docs
	"payment-service/internal/repository/postgres"
	transportHTTP "payment-service/internal/transport/http"
	"payment-service/internal/usecase"

	_ "github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "210624"),
		getEnv("DB_NAME", "payments_db"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	log.Println("Connected to PostgreSQL (payments_db)")

	repo := postgres.NewPaymentRepository(db)
	uc := usecase.NewPaymentUseCase(repo)
	handler := transportHTTP.NewHandler(uc)
	router := transportHTTP.NewRouter(handler)

	port := getEnv("PORT", "8081")
	log.Printf("Payment Service listening on :%s", port)
	log.Printf("Swagger UI: http://localhost:%s/swagger/index.html", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("run server: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
