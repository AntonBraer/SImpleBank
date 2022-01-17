package api

import (
	"database/sql"
	"net/http"
	db "simplebank/db/sqlc"
	"simplebank/util"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type createUsertRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type UserResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func newUserResponse(user db.User) UserResponse {
	return UserResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
}

// @Summary      CreateUser
// @Tags         Users
// @ID           create-user
// @Description  Create new user
// @Accept       json
// @Produce      json
// @Param        input  body      createUsertRequest  true  "user info"
// @Success      200    {object}  db.User
// @Failure      400    {object}  errorResponse
// @Failure      403    {object}  errorResponse
// @Failure      500    {object}  errorResponse
// @Router       /users [post]
func (server *Server) createUser(ctx *gin.Context) {
	var req createUsertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		NewError(ctx, http.StatusBadRequest, err)
		return
	}
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		NewError(ctx, http.StatusInternalServerError, err)
		return
	}

	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Email:          req.Email,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				NewError(ctx, http.StatusForbidden, err)
				return
			}
		}
		NewError(ctx, http.StatusInternalServerError, err)
		return
	}

	resp := newUserResponse(user)

	ctx.JSON(http.StatusOK, resp)
}

type loginUsertRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	AccessToken string       `json:"access_token"`
	User        UserResponse `json:"user"`
}

// @Summary      LoginUser
// @Tags         Users
// @ID           login-user
// @Description  Login User
// @Accept       json
// @Produce      json
// @Param        input  body      loginUsertRequest  true  "login info"
// @Success      200    {object}  loginUserResponse
// @Failure      400    {object}  errorResponse
// @Failure      401    {object}  errorResponse
// @Failure      404    {object}  errorResponse
// @Failure      500    {object}  errorResponse
// @Router       /users/login [post]
func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUsertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		NewError(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			NewError(ctx, http.StatusNotFound, err)
			return
		}
		NewError(ctx, http.StatusInternalServerError, err)
		return
	}

	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		NewError(ctx, http.StatusUnauthorized, err)
		return
	}

	accessToken, err := server.tokenMaker.CreateToken(user.Username, server.config.AccessTokenDuration)
	if err != nil {
		NewError(ctx, http.StatusInternalServerError, err)
		return
	}

	res := loginUserResponse{
		AccessToken: accessToken,
		User:        newUserResponse(user),
	}

	ctx.JSON(http.StatusOK, res)
}
