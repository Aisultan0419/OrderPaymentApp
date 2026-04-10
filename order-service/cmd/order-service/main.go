// @title           Order Service API
// @version         1.0
// @description     Manages customer orders. Calls Payment Service via REST to authorize payments.
// @host            localhost:8080
// @BasePath        /

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "order-service/docs"

	_ "github.com/lib/pq"

	"order-service/internal/client"
	"order-service/internal/repository/postgres"
	transportHTTP "order-service/internal/transport/http"
	"order-service/internal/usecase"
)

func main() {

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "210624"),
		getEnv("DB_NAME", "orders_db"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	log.Println("Connected to PostgreSQL (orders_db)")

	repo := postgres.NewOrderRepository(db)

	paymentServiceURL := getEnv("PAYMENT_SERVICE_URL", "http://localhost:8081")
	paymentClient := client.NewPaymentServiceClient(paymentServiceURL)

	uc := usecase.NewOrderUseCase(repo, paymentClient)
	handler := transportHTTP.NewHandler(uc)
	router := transportHTTP.NewRouter(handler)

	port := getEnv("PORT", "8080")
	log.Printf("Order Service listening on :%s", port)
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
