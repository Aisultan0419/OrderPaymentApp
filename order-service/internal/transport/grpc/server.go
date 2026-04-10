package grpc

import (
	"log"
	"time"

	pb "github.com/Aisultan0419/ap2-gen/order"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"order-service/internal/usecase"
)

type OrderGRPCServer struct {
	pb.UnimplementedOrderServiceServer
	uc *usecase.OrderUseCase
}

func NewOrderGRPCServer(uc *usecase.OrderUseCase) *OrderGRPCServer {
	return &OrderGRPCServer{uc: uc}
}

func (s *OrderGRPCServer) SubscribeToOrderUpdates(req *pb.OrderRequest, stream pb.OrderService_SubscribeToOrderUpdatesServer) error {
	if req.OrderId == "" {
		return status.Error(codes.InvalidArgument, "order_id is required")
	}

	lastStatus := ""

	for {
		select {
		case <-stream.Context().Done():
			return nil
		default:
		}

		order, err := s.uc.GetOrder(stream.Context(), req.OrderId)
		if err != nil {
			return status.Error(codes.NotFound, "order not found")
		}

		if order.Status != lastStatus {
			lastStatus = order.Status
			if err := stream.Send(&pb.OrderStatusUpdate{
				OrderId: order.ID,
				Status:  order.Status,
			}); err != nil {
				log.Printf("[stream] send error: %v", err)
				return err
			}
		}

		if order.Status == "Paid" || order.Status == "Failed" || order.Status == "Cancelled" {
			return nil
		}

		time.Sleep(1 * time.Second)
	}
}
