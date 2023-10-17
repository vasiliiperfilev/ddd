package main

import (
	"context"

	"github.com/vasiliiperfilev/ddd/internal/common/genproto/trainer"
	"github.com/vasiliiperfilev/ddd/internal/trainer/domain/hour"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcServer struct {
	hourRepository hour.Repository
	trainer.UnimplementedTrainerServiceServer
}

func (g GrpcServer) MakeHourAvailable(ctx context.Context, request *trainer.UpdateHourRequest) (*trainer.EmptyResponse, error) {
	if err := g.hourRepository.UpdateHour(ctx, request.Time.AsTime(), func(h *hour.Hour) (*hour.Hour, error) {
		if err := h.MakeAvailable(); err != nil {
			return nil, err
		}

		return h, nil
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &trainer.EmptyResponse{}, nil
}

func (g GrpcServer) ScheduleTraining(ctx context.Context, request *trainer.UpdateHourRequest) (*trainer.EmptyResponse, error) {
	if err := g.hourRepository.UpdateHour(ctx, request.Time.AsTime(), func(h *hour.Hour) (*hour.Hour, error) {
		if err := h.ScheduleTraining(); err != nil {
			return nil, err
		}
		return h, nil
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &trainer.EmptyResponse{}, nil
}

func (g GrpcServer) CancelTraining(ctx context.Context, request *trainer.UpdateHourRequest) (*trainer.EmptyResponse, error) {
	if err := g.hourRepository.UpdateHour(ctx, request.Time.AsTime(), func(h *hour.Hour) (*hour.Hour, error) {
		if err := h.CancelTraining(); err != nil {
			return nil, err
		}
		return h, nil
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &trainer.EmptyResponse{}, nil
}

func (g GrpcServer) IsHourAvailable(ctx context.Context, request *trainer.IsHourAvailableRequest) (*trainer.IsHourAvailableResponse, error) {
	h, err := g.hourRepository.GetOrCreateHour(ctx, request.Time.AsTime())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &trainer.IsHourAvailableResponse{IsAvailable: h.IsAvailable()}, nil
}
