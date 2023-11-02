package main

import (
	"context"
	"flag"
	"github.com/grepplabs/kafka-proxy/pkg/apis"
	"github.com/grepplabs/kafka-proxy/plugin/token-provider/shared"
	"github.com/hashicorp/go-plugin"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/jws"
	"os"
)

const (
	StatusOK            = 0
	StatusEncodeError   = 1
	StatusFileReadError = 2
	AlgorithmNone       = "none"
	jwtFilePath         = "/var/run/secrets/tokens/service-token"
)

type UnsecuredJWTProvider struct {
	claimSub string
}

func (v UnsecuredJWTProvider) GetToken(ctx context.Context, request apis.TokenRequest) (apis.TokenResponse, error) {
	token, err := os.ReadFile(jwtFilePath)
	if err != nil {
		os.Exit(StatusFileReadError)
	}
	return getGetTokenResponse(StatusOK, string(token))
}

func getGetTokenResponse(status int, token string) (apis.TokenResponse, error) {
	success := status == StatusOK
	logrus.Infof("MATTEO DEBUGGING: GetTokenResponse: success=%v, status=%d, token=%s", success, status, token)
	return apis.TokenResponse{Success: success, Status: int32(status), Token: token}, nil
}

func (v UnsecuredJWTProvider) encodeToken() (string, error) {
	header := &jws.Header{
		Algorithm: AlgorithmNone,
	}
	claims := &jws.ClaimSet{
		Sub: v.claimSub,
	}
	signer := func(data []byte) (sig []byte, err error) {
		return []byte{}, nil
	}
	return jws.EncodeWithSigner(header, claims, signer)
}

type pluginMeta struct {
	claimSub string
}

func (f *pluginMeta) flagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("unsecured-jwt-info info settings", flag.ContinueOnError)
	fs.StringVar(&f.claimSub, "claim-sub", "", "subject claim")
	return fs
}

func main() {
	pluginMeta := &pluginMeta{}
	flags := pluginMeta.flagSet()
	_ = flags.Parse(os.Args[1:])

	if pluginMeta.claimSub == "" {
		logrus.Infof("parameter claim-sub is not set. But that is okay for this plugin.")
	}

	unsecuredJWTProvider := &UnsecuredJWTProvider{
		claimSub: pluginMeta.claimSub,
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins: map[string]plugin.Plugin{
			"unsecuredJWTProvider": &shared.TokenProviderPlugin{Impl: unsecuredJWTProvider},
		},
		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
