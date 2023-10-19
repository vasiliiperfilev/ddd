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
	"github.com/vasiliiperfilev/ddd/internal/users/infra"
	"github.com/vasiliiperfilev/ddd/internal/users/oapi"
	"google.golang.org/grpc"
)

func main() {
	logs.Init()

	ctx := context.Background()
	firestoreClient, err := firestore.NewClient(ctx, os.Getenv("GCP_PROJECT"))
	if err != nil {
		panic(err)
	}
	userRepository := infra.NewFirestoreUserRepository(firestoreClient)

	serverType := strings.ToLower(os.Getenv("SERVER_TO_RUN"))
	switch serverType {
	case "http":
		go loadFixtures()

		err := server.RunHTTPServer(func(router chi.Router) server.HandlerWithBackgroundJobs {
			srv := HttpServer{userRepository, &server.BackgroundJobs{}}
			return server.HandlerWithBackgroundJobs{
				Handler:        oapi.HandlerFromMux(srv, router),
				BackgroundJobs: srv.backgroundJobs,
			}
		})
		if err != nil {
			logrus.Fatal(err)
		}
	case "grpc":
		server.RunGRPCServer(func(server *grpc.Server) {
			svc := GrpcServer{userRepository, users.UnimplementedUsersServiceServer{}}
			users.RegisterUsersServiceServer(server, svc)
		})
	default:
		panic(fmt.Sprintf("server type '%s' is not supported", serverType))
	}
}
