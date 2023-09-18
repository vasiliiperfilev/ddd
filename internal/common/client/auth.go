package client

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/api/idtoken"
	"google.golang.org/grpc/credentials"
)

type metadataServerToken struct {
	serviceURL string
}

func newMetadataServerToken(grpcAddr string) credentials.PerRPCCredentials {
	// based on https://cloud.google.com/run/docs/authenticating/service-to-service#go
	// service need to have https prefix without port
	serviceURL := "https://" + strings.Split(grpcAddr, ":")[0]

	return metadataServerToken{serviceURL}
}

// GetRequestMetadata is called on every request, so we are sure that token is always not expired
func (t metadataServerToken) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	ts, err := idtoken.NewTokenSource(ctx, t.serviceURL)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create token source for gRPC")
	}
	idToken, err := ts.Token()
	if err != nil {
		return nil, errors.Wrap(err, "cannot query id token for gRPC")
	}

	return map[string]string{
		"authorization": "Bearer " + idToken.AccessToken,
	}, nil
}

func (metadataServerToken) RequireTransportSecurity() bool {
	return true
}
