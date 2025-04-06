package server

import (
	"booking-service/internal/interceptors"
	"booking-service/internal/logger"
	"booking-service/internal/models"
	"booking-service/internal/service"
	"context"

	"github.com/google/uuid"
	bookingpb "github.com/pet-booking-system/proto-definitions/booking"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BookingHandler struct {
	bookingpb.UnimplementedBookingServiceServer
	service service.BookingService
}

func NewBookingHandler(svc service.BookingService) *BookingHandler {
	return &BookingHandler{service: svc}
}

func (h *BookingHandler) CreateBooking(ctx context.Context, req *bookingpb.CreateBookingRequest) (*bookingpb.CreateBookingResponse, error) {
	logger.Info("CreateBooking called with request: ", req)
	userIDStr, ok := ctx.Value(interceptors.UserIDKey).(string)
	if !ok || userIDStr == "" {
		logger.Error("userID not found in context")
		return nil, status.Errorf(codes.Internal, "userID not found in context")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.Error("invalid userID: ", err)
		return nil, status.Errorf(codes.Internal, "invalid userID: %v", err)
	}

	resourceID, err := uuid.Parse(req.GetResourceId())
	if err != nil {
		logger.Error("invalid resource_id: ", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid resource_id: %v", err)
	}

	start := req.GetStartTime().AsTime().UTC()
	end := req.GetEndTime().AsTime().UTC()
	if end.Before(start) {
		logger.Error("end_time before start_time")
		return nil, status.Error(codes.InvalidArgument, "end_time must be after start_time")
	}

	input := models.CreateBookingInput{
		UserID:     userID,
		ResourceID: resourceID,
		StartTime:  start,
		EndTime:    end,
	}

	booking, err := h.service.Create(ctx, input)
	if err != nil {
		logger.Error("failed to create booking: ", err)
		return nil, status.Errorf(codes.Internal, "failed to create booking: %v", err)
	}

	logger.Info("Booking created: ", booking.ID)
	return &bookingpb.CreateBookingResponse{
		BookingId: booking.ID.String(),
		Status:    booking.Status,
	}, nil
}

func (h *BookingHandler) CancelBooking(ctx context.Context, req *bookingpb.CancelBookingRequest) (*bookingpb.CancelBookingResponse, error) {
	logger.Info("CancelBooking called with request: ", req)
	userIDStr, ok := ctx.Value(interceptors.UserIDKey).(string)
	if !ok || userIDStr == "" {
		logger.Error("userID not found in context")
		return nil, status.Error(codes.Internal, "userID not found in context")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.Error("invalid userID: ", err)
		return nil, status.Errorf(codes.Internal, "invalid userID: %v", err)
	}

	bookingID, err := uuid.Parse(req.GetBookingId())
	if err != nil {
		logger.Error("invalid booking_id: ", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid booking_id: %v", err)
	}

	if err := h.service.Cancel(ctx, bookingID, userID); err != nil {
		logger.Error("failed to cancel booking: ", err)
		return nil, err
	}

	logger.Info("Booking canceled: ", bookingID)
	return &bookingpb.CancelBookingResponse{
		Status: models.StatusCanceled,
	}, nil
}

func (h *BookingHandler) GetBookingStatus(ctx context.Context, req *bookingpb.GetBookingStatusRequest) (*bookingpb.GetBookingStatusResponse, error) {
	logger.Info("GetBookingStatus called with request: ", req)
	bookingID, err := uuid.Parse(req.GetBookingId())
	if err != nil {
		logger.Error("invalid booking_id: ", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid booking_id: %v", err)
	}

	statusStr, err := h.service.GetStatus(ctx, bookingID)
	if err != nil {
		logger.Error("failed to get booking status: ", err)
		return nil, err
	}

	logger.Info("Fetched status for booking ", bookingID, ": ", statusStr)
	return &bookingpb.GetBookingStatusResponse{
		Status: statusStr,
	}, nil
}
