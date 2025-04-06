package models

import (
	"time"

	"github.com/google/uuid"
)

type Booking struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	ResourceID uuid.UUID
	StartTime  time.Time
	EndTime    time.Time
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

const (
	StatusPendingPayment = "pending_payment"
	StatusReserved       = "reserved"
	StatusPaid           = "paid"
	StatusCanceled       = "canceled"
)

type CreateBookingInput struct {
	UserID     uuid.UUID
	ResourceID uuid.UUID
	StartTime  time.Time
	EndTime    time.Time
}
