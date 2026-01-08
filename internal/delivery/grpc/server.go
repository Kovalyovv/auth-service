package grpc

import (
	"context"
	"errors"

	"github.com/Kovalyovv/auth-service/internal/domain"
	"github.com/Kovalyovv/auth-service/internal/usecase"
	"github.com/Kovalyovv/auth-service/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		if errors.Is(err, domain.ErrTokenExpired) {
			return nil, status.Error(codes.Unauthenticated, "token has expired")
		}
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	return &pb.VerifyTokenResponse{
		UserId: userID,
		Valid:  true,
	}, nil
}
