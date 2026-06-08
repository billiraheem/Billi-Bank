package api

import (
	"fmt"

	db "github.com/billiraheem/Billi-Bank/db/sqlc"
	"github.com/billiraheem/Billi-Bank/token"
	"github.com/billiraheem/Billi-Bank/utils"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Server serves HTTP requests for the banking service.
type Server struct {
	config utils.Config
	store  db.Store
	tokenMaker token.Maker
	router *gin.Engine
}

// NewServer creates a new HTTP server and setup routing
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

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	// endpoints
	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccount)
	router.PATCH("/accounts/:id", server.updateAccount)
	router.DELETE("/accounts/:id", server.deleteAccount)

	router.POST("/transfers", server.createTransfer)

	server.router = router
}

// Start runs the HTTP server on a specific address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errRes(err error) gin.H {
	return gin.H{"error": err.Error()}
}