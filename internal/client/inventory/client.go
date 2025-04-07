package inventory

import (
	"context"

	inventorypb "github.com/pet-booking-system/proto-definitions/inventory"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Client struct {
	conn   *grpc.ClientConn
	client inventorypb.InventoryServiceClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}
	return &Client{
		conn:   conn,
		client: inventorypb.NewInventoryServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) CheckAvailability(ctx context.Context, resourceID string) (bool, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false, status.Error(codes.Unauthenticated, "missing metadata")
	}

	ctxWithMeta := metadata.NewOutgoingContext(context.Background(), md)

	resp, err := c.client.CheckAvailability(ctxWithMeta, &inventorypb.CheckAvailabilityRequest{
		ResourceId: resourceID,
	})
	if err != nil {
		return false, err
	}
	return resp.GetIsAvailable(), nil
}

func (c *Client) UpdateStatus(ctx context.Context, resourceID, newStatus string) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "missing metadata")
	}

	ctxWithMeta := metadata.NewOutgoingContext(context.Background(), md)

	_, err := c.client.UpdateResourceStatus(ctxWithMeta, &inventorypb.UpdateResourceStatusRequest{
		ResourceId: resourceID,
		NewStatus:  newStatus,
	})
	return err
}
