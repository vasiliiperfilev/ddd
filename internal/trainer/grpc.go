package main

import (
	"context"
	"fmt"

	"github.com/vasiliiperfilev/ddd/internal/common/genproto/trainer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcServer struct {
	db db
}

func (g GrpcServer) UpdateHour(ctx context.Context, req *trainer.UpdateHourRequest) (*trainer.EmptyResponse, error) {
	trainingTime := req.Time.AsTime()

	date, err := g.db.DateModel(ctx, trainingTime)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to get data model: %s", err))
	}

	hour, found := date.FindHourInDate(trainingTime)
	if !found {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("%s hour not found in schedule", trainingTime))
	}

	if req.HasTrainingScheduled && !hour.Available {
		return nil, status.Error(codes.FailedPrecondition, "hour is not available for training")
	}

	if req.Available && req.HasTrainingScheduled {
		return nil, status.Error(codes.FailedPrecondition, "cannot set hour as available when it have training scheduled")
	}
	if !req.Available && !req.HasTrainingScheduled {
		return nil, status.Error(codes.FailedPrecondition, "cannot set hour as unavailable when it have no training scheduled")
	}
	hour.Available = req.Available

	if hour.HasTrainingScheduled && hour.HasTrainingScheduled == req.HasTrainingScheduled {
		return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("hour HasTrainingScheduled is already %t", hour.HasTrainingScheduled))
	}

	hour.HasTrainingScheduled = req.HasTrainingScheduled
	if err := g.db.SaveModel(ctx, date); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to save date: %s", err))
	}

	return &trainer.EmptyResponse{}, nil
}

func (g GrpcServer) IsHourAvailable(ctx context.Context, req *trainer.IsHourAvailableRequest) (*trainer.IsHourAvailableResponse, error) {
	timeToCheck := req.Time.AsTime()
	model, err := g.db.DateModel(ctx, req.Time.AsTime())
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to get data model: %s", err))
	}

	if hour, found := model.FindHourInDate(timeToCheck); found {
		return &trainer.IsHourAvailableResponse{IsAvailable: hour.Available && !hour.HasTrainingScheduled}, nil
	}

	return &trainer.IsHourAvailableResponse{IsAvailable: false}, nil
}
