package inventory

import (
	"context"

	inventorypb "github.com/pet-booking-system/proto-definitions/inventory"
	"google.golang.org/grpc"
)

type Client struct {
	conn   *grpc.ClientConn
	client inventorypb.InventoryServiceClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure()) // для dev-среды
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
	resp, err := c.client.CheckAvailability(ctx, &inventorypb.CheckAvailabilityRequest{
		ResourceId: resourceID,
	})
	if err != nil {
		return false, err
	}
	return resp.GetIsAvailable(), nil
}

func (c *Client) UpdateStatus(ctx context.Context, resourceID, status string) error {
	_, err := c.client.UpdateResourceStatus(ctx, &inventorypb.UpdateResourceStatusRequest{
		ResourceId: resourceID,
		NewStatus:  status,
	})
	return err
}
