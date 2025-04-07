package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"net"

	"booking-service/config"
	"booking-service/internal/client/inventory"
	"booking-service/internal/client/payment"
	"booking-service/internal/interceptors"
	"booking-service/internal/logger"
	"booking-service/internal/repository"
	"booking-service/internal/server"
	"booking-service/internal/service"

	_ "github.com/lib/pq"
	bookingpb "github.com/pet-booking-system/proto-definitions/booking"
	"google.golang.org/grpc"
)

func Run() {
	logger.Init()

	cfg := config.Load()

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("unable to ping db: %v", err)
	}

	invClient, err := inventory.NewClient(cfg.InventoryServiceAddr)
	if err != nil {
		log.Fatalf("failed to connect to inventory service: %v", err)
	}
	defer invClient.Close()

	paymentClient, err := payment.NewClient(cfg.PaymentServiceAddr)
	if err != nil {
		log.Fatalf("failed to connect to payment service: %v", err)
	}
	defer paymentClient.Close()

	repo := repository.NewBookingRepository(db)
	svc := service.NewBookingService(repo, invClient, *paymentClient)
	handler := server.NewBookingHandler(svc)

	addr := fmt.Sprintf(":%s", cfg.GRPCPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.AuthInterceptor()),
	)

	bookingpb.RegisterBookingServiceServer(grpcServer, handler)

	logger.Info("BookingService listening on ", addr)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
