package main

import (
	"database/sql"
	"log"
	"net"

	_ "github.com/lib/pq" // PostgreSQL driver
	"google.golang.org/grpc"

	delivery "inventory-service/internal/delivery/grpc"
	postgres "inventory-service/internal/repository"
	"inventory-service/internal/usecase"

	pb "github.com/B4rt0n1/final_proto/gen/go/inventory/v1"
)

func main() {
	// 1. Initialize DB Connection
	db, err := sql.Open("postgres", "postgres://user:pass@localhost:5432/inventory?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}
	defer db.Close()

	// 2. Initialize Clean Architecture Layers (Dependency Injection)
	repo := postgres.NewPostgreCarRepository(db)
	carUsecase := usecase.NewCarUsecase(repo)
	grpcHandler := delivery.NewCarInventoryHandler(carUsecase)

	// 3. Start gRPC Server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	// Register the handler with the generated gRPC server interface
	pb.RegisterCarInventoryServiceServer(grpcServer, grpcHandler)

	log.Println("Starting gRPC server on port :50051...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
