package main

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vasiliiperfilev/ddd/internal/common/genproto/users"
	"github.com/vasiliiperfilev/ddd/internal/users/domain/user"
)

type GrpcServer struct {
	userRepository user.Repository
	users.UnimplementedUsersServiceServer
}

func (g GrpcServer) GetTrainingBalance(ctx context.Context, request *users.GetTrainingBalanceRequest) (*users.GetTrainingBalanceResponse, error) {
	user, err := g.userRepository.Get(ctx, request.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &users.GetTrainingBalanceResponse{Amount: int64(user.Balance)}, nil
}

func (g GrpcServer) UpdateTrainingBalance(
	ctx context.Context,
	req *users.UpdateTrainingBalanceRequest,
) (*users.EmptyResponse, error) {
	err := g.userRepository.Update(ctx, req.UserId, func(u user.User) (user.User, error) {
		err := u.UpdateBalance(int(req.AmountChange))
		if err != nil {
			return user.User{}, err
		}
		return u, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update balance: %s", err))
	}

	return &users.EmptyResponse{}, nil
}
