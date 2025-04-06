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

type bookingService struct {
	repo repository.BookingRepository
}

func NewBookingService(repo repository.BookingRepository) BookingService {
	return &bookingService{repo: repo}
}

func (s *bookingService) Create(ctx context.Context, input models.CreateBookingInput) (*models.Booking, error) {
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

	err := s.repo.CreateBooking(ctx, booking)
	if err != nil {
		logger.Error("Failed to create booking in repo: ", err)
		return nil, err
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

	logger.Info("Booking canceled: ", bookingID)
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
