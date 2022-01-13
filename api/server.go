package api

import (
	db "simplebank/db/sqlc"

	"github.com/gin-gonic/gin"
)

type Server struct {
	store  db.Store
	router *gin.Engine
}

// NewServer creates HTTP servre and setup routes
func NewServer(store db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccount)

	server.router = router
	return server
}

func (server *Server) Start(addres string) error {
	return server.router.Run(addres)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
