package grpc

import (
	"context"
	"time"

	pb "github.com/Aisultan0419/ap2-gen/payment"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"payment-service/internal/usecase"
)

type PaymentGRPCServer struct {
	pb.UnimplementedPaymentServiceServer
	uc *usecase.PaymentUseCase
}

func NewPaymentGRPCServer(uc *usecase.PaymentUseCase) *PaymentGRPCServer {
	return &PaymentGRPCServer{uc: uc}
}

func (s *PaymentGRPCServer) ProcessPayment(ctx context.Context, req *pb.PaymentRequest) (*pb.PaymentResponse, error) {
	if req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}
	if req.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be greater than 0")
	}

	payment, err := s.uc.Authorize(ctx, usecase.AuthorizeInput{
		OrderID: req.OrderId,
		Amount:  req.Amount,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.PaymentResponse{
		TransactionId: payment.TransactionID,
		Status:        payment.Status,
		ProcessedAt:   timestamppb.New(time.Now().UTC()),
	}, nil
}
