package main

import (
	"database/sql"
	"log"
	"net"
	"time"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"

	delivery "inventory-service/internal/delivery/grpc"
	infraNats "inventory-service/internal/infrastructure/nats"
	"inventory-service/internal/repository/postgres"
	repoRedis "inventory-service/internal/repository/redis"
	"inventory-service/internal/usecase"

	pb "github.com/B4rt0n1/final_proto/gen/go/inventory/v1"
)

func main() {
	// 1. Database connection init
	db, err := sql.Open("postgres", "postgres://user:pass@localhost:5432/inventory?sslmode=disable")
	if err != nil {
		log.Fatalf("failed Postgres connection execution: %v", err)
	}
	defer db.Close()

	// 2. Redis cluster instance allocation allocation
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer rdb.Close()

	// 3. Connection pipeline implementation out to NATS server
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf("failed execution setting up NATS connection: %v", err)
	}
	defer nc.Close()

	// 4. Inject structural drivers down into structural boundaries
	repo := postgres.NewPostgreCarRepository(db)
	cache := repoRedis.NewRedisCarCache(rdb, 15*time.Minute)
	publisher := infraNats.NewNatsEventPublisher(nc)

	// UseCase encapsulates dependencies across business domain entities
	carUsecase := usecase.NewCarUsecase(repo, cache, publisher)
	grpcHandler := delivery.NewCarInventoryHandler(carUsecase)

	// 5. Fire up the execution network port handler standard context
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed socket allocation target: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCarInventoryServiceServer(grpcServer, grpcHandler)

	log.Println("Server executing cleanly running on port :50051...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to complete gRPC payload loop cycles: %v", err)
	}
}
