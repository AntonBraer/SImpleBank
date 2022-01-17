package api

import (
	"database/sql"
	"errors"
	"net/http"
	db "simplebank/db/sqlc"
	"simplebank/token"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type createAccountRequest struct {
	Currency string `json:"currency" binding:"required,currency"`
}

// @Summary      CreateAccount
// @Security     ApiKeyAuth
// @Tags         Account
// @ID           create-account
// @Description  Create new account
// @Accept       json
// @Produce      json
// @Param        input  body      createAccountRequest  true  "currency"
// @Success      200    {object}  db.Account
// @Failure      400    {object}  errorResponse
// @Failure      403    {object}  errorResponse
// @Failure      500    {object}  errorResponse
// @Router       /accounts [post]
func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		NewError(ctx, http.StatusBadRequest, err)
		return
	}

	authPayload := ctx.MustGet(authPayloadKey).(*token.Payload)

	arg := db.CreateAccountParams{
		Owner:    authPayload.Username,
		Balance:  0,
		Currency: req.Currency,
	}

	account, err := server.store.CreateAccount(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation", "foreign_key_violation":
				NewError(ctx, http.StatusForbidden, err)
				return
			}
		}
		NewError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, account)
}

type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// @Summary      GetAccount
// @Security     ApiKeyAuth
// @Tags         Account
// @ID           get-account
// @Description  Get account
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Account ID"
// @Success      200  {object}  db.Account
// @Failure      400  {object}  errorResponse
// @Failure      401  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /accounts/{id} [get]
func (server *Server) getAccount(ctx *gin.Context) {
	var req getAccountRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		NewError(ctx, http.StatusBadRequest, err)
		return
	}

	account, err := server.store.GetAccount(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			NewError(ctx, http.StatusNotFound, err)
			return
		}
		NewError(ctx, http.StatusInternalServerError, err)
		return
	}
	authPayload := ctx.MustGet(authPayloadKey).(*token.Payload)
	if authPayload.Username != account.Owner {
		err := errors.New("account doesn't belong to the authenticated user")
		NewError(ctx, http.StatusUnauthorized, err)
		return
	}
	ctx.JSON(http.StatusOK, account)
}

type listAccountRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

// @Summary      ListAccount
// @Security     ApiKeyAuth
// @Tags         Account
// @ID           list-account
// @Description  List account
// @Accept       json
// @Produce      json
// @Param        page_id    query     int  false  "Page ID"
// @Param        page_size  query     int  false  "Page Size"
// @Success      200        {array}   db.Account
// @Failure      400        {object}  errorResponse
// @Failure      500        {object}  errorResponse
// @Router       /accounts [get]
func (server *Server) listAccount(ctx *gin.Context) {
	var req listAccountRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		NewError(ctx, http.StatusBadRequest, err)
		return
	}
	authPayload := ctx.MustGet(authPayloadKey).(*token.Payload)

	arg := db.ListAccountsParams{
		Owner:  authPayload.Username,
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}
	account, err := server.store.ListAccounts(ctx, arg)
	if err != nil {
		NewError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, account)
}
