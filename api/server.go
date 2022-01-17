package api

import (
	"fmt"
	db "simplebank/db/sqlc"
	"simplebank/token"
	"simplebank/util"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

type Server struct {
	store      db.Store
	config     util.Config
	tokenMaker token.Maker
	router     *gin.Engine
}

// NewServer creates HTTP servre and setup routes
func NewServer(config util.Config, store db.Store) (*Server, error) {
	maker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create maker: %w", err)
	}
	server := &Server{
		store:      store,
		config:     config,
		tokenMaker: maker,
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	server.createRoutes()

	return server, nil
}

func (server *Server) createRoutes() {
	router := gin.Default()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))

	authRoutes.POST("/accounts", server.createAccount)
	authRoutes.GET("/accounts/:id", server.getAccount)
	authRoutes.GET("/accounts", server.listAccount)
	authRoutes.POST("/transfers", server.createTransfer)

	server.router = router
}

func (server *Server) Start(addres string) error {
	return server.router.Run(addres)
}

type errorResponse struct{
	Code int `json:"code"`
	Message string `json:"error"`
}

func NewError(ctx *gin.Context, status int, err error) {
	er := errorResponse{
		Code: status,
		Message: err.Error(),
	}
	ctx.JSON(status, er)
}
