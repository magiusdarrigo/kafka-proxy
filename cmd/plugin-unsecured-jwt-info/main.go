package main

import (
	"context"
	"github.com/grepplabs/kafka-proxy/pkg/apis"
	"github.com/grepplabs/kafka-proxy/plugin/token-provider/shared"
	"github.com/hashicorp/go-plugin"
	"github.com/sirupsen/logrus"
	"os"
)

const (
	StatusOK            = 0
	StatusEncodeError   = 1
	StatusFileReadError = 2
	jwtFilePath         = "/var/run/secrets/tokens/service-token"
)

type UnsecuredJWTProvider struct{}

func (v UnsecuredJWTProvider) GetToken(ctx context.Context, request apis.TokenRequest) (apis.TokenResponse, error) {
	token, err := os.ReadFile(jwtFilePath)
	if err != nil {
		logrus.Errorf("Error reading JWT from file: %v", err)
		return getGetTokenResponse(StatusFileReadError, "")
	}
	return getGetTokenResponse(StatusOK, string(token))
}

func getGetTokenResponse(status int, token string) (apis.TokenResponse, error) {
	success := status == StatusOK
	return apis.TokenResponse{Success: success, Status: int32(status), Token: token}, nil
}

func main() {
	unsecuredJWTProvider := &UnsecuredJWTProvider{}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins: map[string]plugin.Plugin{
			"unsecuredJWTProvider": &shared.TokenProviderPlugin{Impl: unsecuredJWTProvider},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
