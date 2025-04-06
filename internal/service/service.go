package service

import (
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
	Create(ctx context.Context, input models.CreateBookingInput) (*models.Booking, error)
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
}

func NewBookingService(repo repository.BookingRepository, inventory InventoryClient) BookingService {
	return &bookingService{
		repo:      repo,
		inventory: inventory,
	}
}

func (s *bookingService) Create(ctx context.Context, input models.CreateBookingInput) (*models.Booking, error) {
	ok, err := s.inventory.CheckAvailability(ctx, input.ResourceID.String())
	if err != nil {
		logger.Error("Inventory check failed: ", err)
		return nil, status.Errorf(codes.Internal, "inventory check failed: %v", err)
	}
	if !ok {
		logger.Info("Resource not available for booking: ", input.ResourceID)
		return nil, status.Error(codes.FailedPrecondition, "resource not available")
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
		return nil, err
	}

	if err := s.inventory.UpdateStatus(ctx, input.ResourceID.String(), "booked"); err != nil {
		logger.Error("Failed to update resource status to booked: ", err)
		return nil, status.Errorf(codes.Internal, "inventory update failed: %v", err)
	}

	logger.Info("Created booking: ", booking.ID)
	return booking, nil
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
