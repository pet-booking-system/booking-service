package repository

import (
	"booking-service/internal/logger"
	"booking-service/internal/models"
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type BookingRepository interface {
	CreateBooking(ctx context.Context, b *models.Booking) error
	GetBookingByID(ctx context.Context, id uuid.UUID) (*models.Booking, error)
	UpdateBookingStatus(ctx context.Context, id uuid.UUID, status string) error
}

type bookingRepo struct {
	db *sql.DB
}

func NewBookingRepository(db *sql.DB) BookingRepository {
	return &bookingRepo{db: db}
}

func (r *bookingRepo) CreateBooking(ctx context.Context, b *models.Booking) error {
	query := `
		INSERT INTO bookings (
			id, user_id, resource_id,
			start_time, end_time,
			status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		b.ID,
		b.UserID,
		b.ResourceID,
		b.StartTime,
		b.EndTime,
		b.Status,
		b.CreatedAt,
		b.UpdatedAt,
	)

	if err != nil {
		logger.Error("DB: failed to insert booking: ", err)
	}
	return err
}

func (r *bookingRepo) GetBookingByID(ctx context.Context, id uuid.UUID) (*models.Booking, error) {
	query := `SELECT id, user_id, resource_id, start_time, end_time, status, created_at, updated_at
	          FROM bookings WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, id)

	var b models.Booking
	err := row.Scan(
		&b.ID, &b.UserID, &b.ResourceID,
		&b.StartTime, &b.EndTime, &b.Status,
		&b.CreatedAt, &b.UpdatedAt,
	)

	if err != nil {
		logger.Error("DB: failed to fetch booking: ", err)
		return nil, err
	}

	return &b, nil
}

func (r *bookingRepo) UpdateBookingStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE bookings SET status = $1, updated_at = now() WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		logger.Error("DB: failed to update booking status: ", err)
		return err
	}

	logger.Info("Booking status updated: ", id, " -> ", status)
	return nil
}
