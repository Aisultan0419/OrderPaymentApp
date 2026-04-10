// @title           Order Service API
// @version         1.0
// @description     Manages customer orders. Calls Payment Service via gRPC to authorize payments.
// @host            localhost:8080
// @BasePath        /

package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	_ "order-service/docs"
	"order-service/internal/client"
	"order-service/internal/repository/postgres"
	transportGRPC "order-service/internal/transport/grpc"
	transportHTTP "order-service/internal/transport/http"
	"order-service/internal/usecase"

	pbOrder "github.com/Aisultan0419/ap2-gen/order"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
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

	paymentGRPCAddr := getEnv("PAYMENT_GRPC_ADDR", "localhost:9091")
	paymentClient, err := client.NewPaymentGRPCClient(paymentGRPCAddr)
	if err != nil {
		log.Fatalf("create payment grpc client: %v", err)
	}

	uc := usecase.NewOrderUseCase(repo, paymentClient)

	// Streaming gRPC server
	grpcPort := getEnv("GRPC_PORT", "9090")
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen grpc: %v", err)
	}
	grpcServer := grpc.NewServer()
	pbOrder.RegisterOrderServiceServer(grpcServer, transportGRPC.NewOrderGRPCServer(uc))

	go func() {
		log.Printf("Order gRPC server listening on :%s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("grpc serve: %v", err)
		}
	}()

	// HTTP server
	handler := transportHTTP.NewHandler(uc)
	router := transportHTTP.NewRouter(handler)

	port := getEnv("PORT", "8080")
	log.Printf("Order HTTP server listening on :%s", port)
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
