package gapi

import (
	"fmt"

	db "github.com/billiraheem/Billi-Bank/db/sqlc"
	"github.com/billiraheem/Billi-Bank/pb"
	"github.com/billiraheem/Billi-Bank/token"
	"github.com/billiraheem/Billi-Bank/utils"
)

// Server serves gRPC requests for the banking service.
type Server struct {
	pb.UnimplementedBilliBankServer
	config utils.Config
	store  db.Store
	tokenMaker token.Maker
}

// NewServer creates a new gRPC server and setup routing
func NewServer(config utils.Config, store db.Store) (*Server, error) {
	// to use JWT: swap NewPasetoMaker with NewJWTMaker
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config: config,
		store: store,
		tokenMaker: tokenMaker,
	}

	return server, nil
}