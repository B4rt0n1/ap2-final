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

	"github.com/B4rt0n1/ap2-final/services/user-service/internal/infrastructure/config"
	usernats "github.com/B4rt0n1/ap2-final/services/user-service/internal/infrastructure/nats"
	userpostgres "github.com/B4rt0n1/ap2-final/services/user-service/internal/infrastructure/postgres"
	grpcserver "github.com/B4rt0n1/ap2-final/services/user-service/internal/transport/grpc"
	"github.com/B4rt0n1/ap2-final/services/user-service/internal/usecase"
	userv1 "github.com/B4rt0n1/final_proto/gen/go/user/v1"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
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

	natsConn, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		return err
	}
	defer natsConn.Drain()

	repo := userpostgres.NewUserRepository(db)
	userService := usecase.NewUserUseCase(repo, natsConn)
	usernats.NewEmailWorker(natsConn, repo, usernats.SMTPConfig{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		Username: cfg.SMTP.Username,
		Password: cfg.SMTP.Password,
		From:     cfg.SMTP.From,
	}).Start()
	handler := grpcserver.NewUserHandler(userService)

	server := grpc.NewServer()
	userv1.RegisterUserServiceServer(server, handler)

	listener, err := net.Listen("tcp", cfg.GRPCAddress)
	if err != nil {
		return err
	}

	serveErr := make(chan error, 1)
	go func() {
		log.Printf("User Service gRPC server listening on %s", cfg.GRPCAddress)
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
