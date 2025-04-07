package payment

import (
	"context"

	paymentpb "github.com/pet-booking-system/proto-definitions/payment"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client paymentpb.PaymentServiceClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   conn,
		client: paymentpb.NewPaymentServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) ProcessPayment(ctx context.Context, bookingID, userID string, amount float64) (*paymentpb.ProcessPaymentResponse, error) {
	return c.client.ProcessPayment(ctx, &paymentpb.ProcessPaymentRequest{
		BookingId: bookingID,
		UserId:    userID,
		Amount:    amount,
	})
}
