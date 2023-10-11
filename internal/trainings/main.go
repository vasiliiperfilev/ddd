package main

import (
	"context"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	grpcClient "github.com/vasiliiperfilev/ddd/internal/common/client"
	"github.com/vasiliiperfilev/ddd/internal/common/logs"
	"github.com/vasiliiperfilev/ddd/internal/common/server"
)

func main() {
	logs.Init()

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, os.Getenv("GCP_PROJECT"))
	if err != nil {
		panic(err)
	}

	trainerClient, closeTrainerClient, err := grpcClient.NewTrainerClient()
	if err != nil {
		panic(err)
	}
	defer closeTrainerClient()

	usersClient, closeUsersClient, err := grpcClient.NewUsersClient()
	if err != nil {
		panic(err)
	}
	defer closeUsersClient()

	firebaseDB := db{client}

	err = server.RunHTTPServer(func(router chi.Router) server.HandlerWithBackgroundJobs {
		srv := HttpServer{firebaseDB, trainerClient, usersClient, &server.BackgroundJobs{}}
		return server.HandlerWithBackgroundJobs{
			Handler:        HandlerFromMux(srv, router),
			BackgroundJobs: srv.backgroundJobs,
		}
	})
	if err != nil {
		logrus.Fatal(err)
	}
}
