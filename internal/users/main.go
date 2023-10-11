package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"github.com/vasiliiperfilev/ddd/internal/common/genproto/users"
	"github.com/vasiliiperfilev/ddd/internal/common/logs"
	"github.com/vasiliiperfilev/ddd/internal/common/server"
	"google.golang.org/grpc"
)

func main() {
	logs.Init()

	ctx := context.Background()
	firestoreClient, err := firestore.NewClient(ctx, os.Getenv("GCP_PROJECT"))
	if err != nil {
		panic(err)
	}
	firebaseDB := db{firestoreClient}

	serverType := strings.ToLower(os.Getenv("SERVER_TO_RUN"))
	switch serverType {
	case "http":
		go loadFixtures(firebaseDB)

		err := server.RunHTTPServer(func(router chi.Router) server.HandlerWithBackgroundJobs {
			srv := HttpServer{firebaseDB, &server.BackgroundJobs{}}
			return server.HandlerWithBackgroundJobs{
				Handler:        HandlerFromMux(srv, router),
				BackgroundJobs: srv.backgroundJobs,
			}
		})
		if err != nil {
			logrus.Fatal(err)
		}
	case "grpc":
		server.RunGRPCServer(func(server *grpc.Server) {
			svc := GrpcServer{firebaseDB}
			users.RegisterUsersServiceServer(server, svc)
		})
	default:
		panic(fmt.Sprintf("server type '%s' is not supported", serverType))
	}
}
