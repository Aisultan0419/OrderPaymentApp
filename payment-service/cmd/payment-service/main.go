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
	"net"
	"os"

	_ "payment-service/docs"
	"payment-service/internal/repository/postgres"
	transportGRPC "payment-service/internal/transport/grpc"
	transportHTTP "payment-service/internal/transport/http"
	"payment-service/internal/usecase"

	pb "github.com/Aisultan0419/ap2-gen/payment"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
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

	// gRPC server
	grpcPort := getEnv("GRPC_PORT", "9091")
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen grpc: %v", err)
	}
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(transportGRPC.LoggingInterceptor),
	)
	pb.RegisterPaymentServiceServer(grpcServer, transportGRPC.NewPaymentGRPCServer(uc))

	go func() {
		log.Printf("Payment gRPC server listening on :%s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("grpc serve: %v", err)
		}
	}()

	// HTTP server
	handler := transportHTTP.NewHandler(uc)
	router := transportHTTP.NewRouter(handler)

	port := getEnv("PORT", "8081")
	log.Printf("Payment HTTP server listening on :%s", port)
	log.Printf("Swagger UI: http://localhost:%s/swagger/index.html", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("run http server: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
