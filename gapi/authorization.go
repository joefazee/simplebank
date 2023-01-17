package gapi

import (
	"context"
	"errors"
	"github.com/joefazee/simplebank/token"
	"google.golang.org/grpc/metadata"
	"strings"
)

const (
	authorization               = "authorization"
	authAuthorizationTypeBearer = "bearer"
)

func (server *Server) authorizeUser(ctx context.Context) (*token.Payload, error) {

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("missing metadata")
	}

	values := md.Get(authorization)
	if len(values) == 0 {
		return nil, errors.New("missing authorization header")
	}

	authHeader := values[0]
	fields := strings.Fields(authHeader)

	if len(fields) < 2 {
		return nil, errors.New("invalid authorization format")
	}

	authType := strings.ToLower(fields[0])
	if authType != authAuthorizationTypeBearer {
		return nil, errors.New("invalid authorization type")
	}

	accessToken := fields[0]

	payload, err := server.tokenMaker.VerifyToken(accessToken)
	if err != nil {
		return nil, errors.New("invalid authorization token")
	}

	return payload, nil

}
