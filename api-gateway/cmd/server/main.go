package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/B4rt0n1/ap2-final/api-gateway/internal/config"
	bookinghttp "github.com/B4rt0n1/ap2-final/api-gateway/internal/transport/http/booking"
	inventoryhttp "github.com/B4rt0n1/ap2-final/api-gateway/internal/transport/http/inventory"
	userhttp "github.com/B4rt0n1/ap2-final/api-gateway/internal/transport/http/user"
	bookingv1 "github.com/B4rt0n1/final_proto/gen/go/booking/v1"
	inventoryv1 "github.com/B4rt0n1/final_proto/gen/go/inventory/v1"
	userv1 "github.com/B4rt0n1/final_proto/gen/go/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const shutdownTimeout = 5 * time.Second

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	bookingConn, err := grpc.NewClient(
		cfg.BookingTarget,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}
	defer bookingConn.Close()

	userConn, err := grpc.NewClient(
		cfg.UserTarget,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}
	defer userConn.Close()

	inventoryConn, err := grpc.NewClient(
		cfg.InventoryTarget,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}
	defer inventoryConn.Close()

	mux := http.NewServeMux()
	bookinghttp.Register(mux, bookingv1.NewBookingServiceClient(bookingConn))
	inventoryhttp.Register(mux, inventoryv1.NewCarInventoryServiceClient(inventoryConn))
	userhttp.Register(mux, userv1.NewUserServiceClient(userConn))
	mux.Handle("/", http.FileServer(http.Dir("api-gateway/web")))

	server := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           mux,
		ReadHeaderTimeout: shutdownTimeout,
	}

	serveErr := make(chan error, 1)
	go func() {
		log.Printf("api gateway listening on %s", cfg.HTTPAddress)
		serveErr <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
