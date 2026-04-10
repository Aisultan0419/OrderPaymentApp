package client

import (
	"context"

	pb "github.com/Aisultan0419/ap2-gen/payment"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"order-service/internal/usecase"
)

type PaymentGRPCClient struct {
	client pb.PaymentServiceClient
}

func NewPaymentGRPCClient(addr string) (usecase.PaymentClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &PaymentGRPCClient{client: pb.NewPaymentServiceClient(conn)}, nil
}

func (c *PaymentGRPCClient) AuthorizePayment(ctx context.Context, orderID string, amount int64) (*usecase.PaymentResult, error) {
	resp, err := c.client.ProcessPayment(ctx, &pb.PaymentRequest{
		OrderId: orderID,
		Amount:  amount,
	})
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.Unavailable {
			return nil, &usecase.ErrServiceUnavailable{Service: "payment"}
		}
		return nil, err
	}
	return &usecase.PaymentResult{
		TransactionID: resp.TransactionId,
		Status:        resp.Status,
	}, nil
}
