package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"github.com/vasiliiperfilev/ddd/internal/common/genproto/trainer"
	"github.com/vasiliiperfilev/ddd/internal/common/logs"
	"github.com/vasiliiperfilev/ddd/internal/common/server"
	"github.com/vasiliiperfilev/ddd/internal/trainer/infra"
	"github.com/vasiliiperfilev/ddd/internal/trainer/openapi_types"
	"google.golang.org/grpc"
)

func main() {
	logs.Init()

	ctx := context.Background()
	firebaseClient, err := firestore.NewClient(ctx, os.Getenv("GCP_PROJECT"))
	if err != nil {
		panic(err)
	}

	firebaseDB := db{firebaseClient}

	serverType := strings.ToLower(os.Getenv("SERVER_TO_RUN"))
	switch serverType {
	case "http":
		go loadFixtures(firebaseDB)

		err := server.RunHTTPServer(func(router chi.Router) server.HandlerWithBackgroundJobs {
			srv := HttpServer{firebaseDB, infra.NewFirestoreHourRepository(firebaseClient), &server.BackgroundJobs{}}
			return server.HandlerWithBackgroundJobs{
				Handler:        openapi_types.HandlerFromMux(srv, router),
				BackgroundJobs: srv.backgroundJobs,
			}
		})
		if err != nil {
			logrus.Fatal(err)
		}
	case "grpc":
		server.RunGRPCServer(func(server *grpc.Server) {
			svc := GrpcServer{infra.NewFirestoreHourRepository(firebaseClient), trainer.UnimplementedTrainerServiceServer{}}
			trainer.RegisterTrainerServiceServer(server, svc)
		})
	default:
		panic(fmt.Sprintf("server type '%s' is not supported", serverType))
	}
}
