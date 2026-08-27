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
	router.POST("/tokens/renew_access", server.renewAccessToken)

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))

	authRoutes.POST("/accounts", server.createAccount)
	authRoutes.GET("/accounts/:id", server.getAccount)
	authRoutes.GET("/accounts", server.listAccount)
	authRoutes.PATCH("/accounts/:id", server.updateAccount)
	authRoutes.DELETE("/accounts/:id", server.deleteAccount)

	authRoutes.PATCH("/users/me", server.updateUser)

	authRoutes.POST("/transfers", server.createTransfer)
	authRoutes.PATCH("/deposit", server.createDeposit)

	server.router = router
}

// Start runs the HTTP server on a specific address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errRes(err error) gin.H {
	return gin.H{"error": err.Error()}
}