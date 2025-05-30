package service

import (
	"booking-service/internal/client/payment"
	"booking-service/internal/logger"
	"booking-service/internal/models"
	"booking-service/internal/repository"
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BookingService interface {
	Create(ctx context.Context, input models.CreateBookingInput) (*models.Booking, string, error)
	Cancel(ctx context.Context, bookingID, userID uuid.UUID) error
	GetStatus(ctx context.Context, bookingID uuid.UUID) (string, error)
}

type InventoryClient interface {
	CheckAvailability(ctx context.Context, resourceID string) (bool, error)
	UpdateStatus(ctx context.Context, resourceID, status string) error
}

type bookingService struct {
	repo      repository.BookingRepository
	inventory InventoryClient
	payment   payment.Client
}

func NewBookingService(repo repository.BookingRepository, inventory InventoryClient, payment payment.Client) BookingService {
	return &bookingService{
		repo:      repo,
		inventory: inventory,
		payment:   payment,
	}
}

func (s *bookingService) Create(ctx context.Context, input models.CreateBookingInput) (*models.Booking, string, error) {
	ok, err := s.inventory.CheckAvailability(ctx, input.ResourceID.String())
	if err != nil {
		logger.Error("Inventory check failed: ", err)
		return nil, "", status.Errorf(codes.Internal, "inventory check failed: %v", err)
	}
	if !ok {
		logger.Info("Resource not available for booking: ", input.ResourceID)
		return nil, "", status.Error(codes.FailedPrecondition, "resource not available")
	}

	id := uuid.New()
	now := time.Now().UTC()

	booking := &models.Booking{
		ID:         id,
		UserID:     input.UserID,
		ResourceID: input.ResourceID,
		StartTime:  input.StartTime,
		EndTime:    input.EndTime,
		Status:     models.StatusPendingPayment,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.repo.CreateBooking(ctx, booking); err != nil {
		logger.Error("Failed to create booking in repo: ", err)
		return nil, "", err
	}

	if err := s.inventory.UpdateStatus(ctx, input.ResourceID.String(), "booked"); err != nil {
		logger.Error("Failed to update resource status to booked: ", err)
		return nil, "", status.Errorf(codes.Internal, "inventory update failed: %v", err)
	}

	resp, err := s.payment.ProcessPayment(ctx, booking.ID.String(), input.UserID.String(), 10000)
	if err != nil {
		logger.Error("Payment processing failed: ", err)
		_ = s.repo.UpdateBookingStatus(ctx, booking.ID, models.StatusCanceled)
		_ = s.inventory.UpdateStatus(ctx, input.ResourceID.String(), "available")
		return nil, "", status.Errorf(codes.Internal, "payment failed, booking canceled")
	}

	if resp.GetStatus() == "failed" {
		logger.Warn("Payment was not successful")
		_ = s.repo.UpdateBookingStatus(ctx, booking.ID, models.StatusCanceled)
		_ = s.inventory.UpdateStatus(ctx, input.ResourceID.String(), "available")
		return nil, "", status.Error(codes.Aborted, "payment unsuccessful, booking canceled")
	}

	logger.Info("Booking created and payment successful: ", booking.ID)
	return booking, resp.GetPaymentId(), nil
}

func (s *bookingService) Cancel(ctx context.Context, bookingID, userID uuid.UUID) error {
	logger.Info("Attempting to cancel booking: ", bookingID)

	booking, err := s.repo.GetBookingByID(ctx, bookingID)
	if err != nil {
		logger.Error("Failed to get booking by ID: ", err)
		return err
	}

	if booking.UserID != userID {
		logger.Info("Permission denied for cancel: booking ", bookingID, ", user ", userID)
		return status.Error(codes.PermissionDenied, "you are not the owner of this booking")
	}

	if err := s.repo.UpdateBookingStatus(ctx, bookingID, models.StatusCanceled); err != nil {
		logger.Error("Failed to update booking status: ", err)
		return err
	}

	if err := s.inventory.UpdateStatus(ctx, booking.ResourceID.String(), "available"); err != nil {
		logger.Error("Failed to update resource status to available: ", err)
		return err
	}

	logger.Info("Booking canceled and resource released: ", bookingID)
	return nil
}

func (s *bookingService) GetStatus(ctx context.Context, bookingID uuid.UUID) (string, error) {
	logger.Info("Fetching status for booking: ", bookingID)

	booking, err := s.repo.GetBookingByID(ctx, bookingID)
	if err != nil {
		logger.Error("Failed to get booking by ID: ", err)
		return "", err
	}

	logger.Info("Status for booking ", bookingID, ": ", booking.Status)
	return booking.Status, nil
}
