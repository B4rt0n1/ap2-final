package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/B4rt0n1/ap2-final/services/booking-service/internal/infrastructure/config"
	"github.com/B4rt0n1/ap2-final/services/booking-service/internal/infrastructure/grpcclient"
	"github.com/B4rt0n1/ap2-final/services/booking-service/internal/infrastructure/id"
	natspublisher "github.com/B4rt0n1/ap2-final/services/booking-service/internal/infrastructure/nats"
	bookingpostgres "github.com/B4rt0n1/ap2-final/services/booking-service/internal/infrastructure/postgres"
	grpcserver "github.com/B4rt0n1/ap2-final/services/booking-service/internal/transport/grpc"
	"github.com/B4rt0n1/ap2-final/services/booking-service/internal/usecase"
	inventoryv1 "github.com/B4rt0n1/final_proto/gen/go/inventory/v1"
	userv1 "github.com/B4rt0n1/final_proto/gen/go/user/v1"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const startupTimeout = 5 * time.Second

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := openDatabase(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	userConn, err := dialRemoteService(ctx, cfg.UserAddress)
	if err != nil {
		return err
	}
	defer userConn.Close()

	inventoryConn, err := dialRemoteService(ctx, cfg.InventoryAddress)
	if err != nil {
		return err
	}
	defer inventoryConn.Close()

	natsConn, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		return err
	}
	defer natsConn.Drain()

	listener, err := net.Listen("tcp", cfg.GRPCAddress)
	if err != nil {
		return err
	}

	repo := bookingpostgres.NewBookingRepository(db)
	inventoryClient := grpcclient.NewInventoryClient(inventoryv1.NewCarInventoryServiceClient(inventoryConn))
	bookingService := usecase.New(usecase.Dependencies{
		Bookings:   repo,
		Transactor: repo,
		IDs:        id.UUIDGenerator{},
		Pricing:    inventoryClient,
		Users:      grpcclient.NewUserClient(userv1.NewUserServiceClient(userConn)),
		Cars:       inventoryClient,
		Events:     natspublisher.New(natsConn),
	})

	server := grpc.NewServer()
	grpcserver.Register(server, bookingService)

	serveErr := make(chan error, 1)
	go func() {
		log.Printf("booking gRPC server listening on %s", cfg.GRPCAddress)
		serveErr <- server.Serve(listener)
	}()

	select {
	case <-ctx.Done():
		server.GracefulStop()
		return nil
	case err := <-serveErr:
		if errors.Is(err, grpc.ErrServerStopped) {
			return nil
		}
		return err
	}
}

func openDatabase(ctx context.Context, databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, startupTimeout)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func dialRemoteService(ctx context.Context, address string) (*grpc.ClientConn, error) {
	dialCtx, cancel := context.WithTimeout(ctx, startupTimeout)
	defer cancel()

	return grpc.DialContext(
		dialCtx,
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}
