package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"github.com/vasiliiperfilev/ddd/internal/common/genproto/trainer"
	"github.com/vasiliiperfilev/ddd/internal/common/logs"
	"github.com/vasiliiperfilev/ddd/internal/common/server"
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

		err := server.RunHTTPServer(func(router chi.Router) http.Handler {
			return HandlerFromMux(HttpServer{firebaseDB}, router)
		})
		if err != nil {
			logrus.Fatal(err)
		}
	case "grpc":
		server.RunGRPCServer(func(server *grpc.Server) {
			svc := GrpcServer{firebaseDB}
			trainer.RegisterTrainerServiceServer(server, svc)
		})
	default:
		panic(fmt.Sprintf("server type '%s' is not supported", serverType))
	}
}
