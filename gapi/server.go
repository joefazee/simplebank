package gapi

import (
	"fmt"
	"github.com/joefazee/simplebank/pb"

	db "github.com/joefazee/simplebank/db/sqlc"
	"github.com/joefazee/simplebank/token"
	"github.com/joefazee/simplebank/util"
)

// Server serves gRPC requests
type Server struct {
	pb.UnimplementedSimpleBankServer
	store      db.Store
	tokenMaker token.Maker
	config     util.Config
}

// NewServer creates new gRPC server.
func NewServer(config util.Config, store db.Store) (*Server, error) {

	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create a token maker")
	}

	server := &Server{
		store:      store,
		tokenMaker: tokenMaker,
		config:     config,
	}

	return server, nil
}
