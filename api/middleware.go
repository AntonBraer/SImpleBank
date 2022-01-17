package api

import (
	"errors"
	"fmt"
	"net/http"
	"simplebank/token"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	authHeaderKey  = "authorization"
	authTypeBearer = "bearer"
	authPayloadKey = "auth_payload"
)

func authMiddleware(token token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader(authHeaderKey)
		if len(authHeader) == 0 {
			err := errors.New("authorization header is not provided")
			NewError(ctx, http.StatusUnauthorized, err)
			return
		}
		fields := strings.Fields(authHeader)
		if len(fields) != 2 {
			err := errors.New("invalid authorization header")
			NewError(ctx, http.StatusUnauthorized, err)
			return
		}
		authType := strings.ToLower(fields[0])
		if authType != authTypeBearer {
			err := fmt.Errorf("unsupported authorization type %s", authType)
			NewError(ctx, http.StatusUnauthorized, err)
			return
		}

		accessToken := fields[1]
		payload, err := token.VerifyToken(accessToken)
		if err != nil {
			NewError(ctx, http.StatusUnauthorized, err)
			return
		}

		ctx.Set(authPayloadKey, payload)
		ctx.Next()
	}
}
