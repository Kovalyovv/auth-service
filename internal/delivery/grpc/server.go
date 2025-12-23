package grpc

import (
	"context"

	"github.com/Kovalyovv/auth-service/internal/usecase"
	"github.com/Kovalyovv/auth-service/pkg/pb"
)

type Server struct {
	pb.UnimplementedAuthServiceServer
	uc *usecase.AuthUseCase
}

func NewServer(uc *usecase.AuthUseCase) *Server {
	return &Server{uc: uc}
}

func (s *Server) VerifyToken(ctx context.Context, req *pb.VerifyTokenRequest) (*pb.VerifyTokenResponse, error) {
	userID, err := s.uc.Verify(req.GetToken())
	if err != nil {
		return &pb.VerifyTokenResponse{Valid: false}, nil
	}

	return &pb.VerifyTokenResponse{
		UserId: userID,
		Valid:  true,
	}, nil
}
