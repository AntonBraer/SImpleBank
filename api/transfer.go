package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	db "simplebank/db/sqlc"
	"simplebank/token"

	"github.com/gin-gonic/gin"
)

type TransferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

// @Summary      CreateTransfer
// @Security     ApiKeyAuth
// @Tags         Transfer
// @ID           create-transfer
// @Description  Create new transfer
// @Accept       json
// @Produce      json
// @Param        input  body      TransferRequest  true  "Transfer info"
// @Success      200    {object}  db.TransferTxResult
// @Failure      400    {object}  errorResponse
// @Failure      401    {object}  errorResponse
// @Failure      500    {object}  errorResponse
// @Router       /transfers [post]
func (server *Server) createTransfer(ctx *gin.Context) {
	var req TransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		NewError(ctx, http.StatusBadRequest, err)
		return
	}

	fromAccount, valid := server.validAccount(ctx, req.FromAccountID, req.Currency)

	if !valid {
		return
	}

	authPayload := ctx.MustGet(authPayloadKey).(*token.Payload)
	if authPayload.Username != fromAccount.Owner {
		err := errors.New("account doesn't belong to the authenticated user")
		NewError(ctx, http.StatusUnauthorized, err)
		return
	}

	_, valid = server.validAccount(ctx, req.ToAccountID, req.Currency)
	if !valid {
		return
	}

	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	result, err := server.store.TransferTx(ctx, arg)
	if err != nil {
		NewError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (server *Server) validAccount(ctx *gin.Context, accountID int64, currency string) (db.Account, bool) {
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			NewError(ctx, http.StatusNotFound, err)
			return account, false
		}
		NewError(ctx, http.StatusInternalServerError, err)
		return account, false
	}
	if account.Currency != currency {
		err := fmt.Errorf("account [%d] currency mistmatch: %s vs %s", accountID, currency, account.Currency)
		NewError(ctx, http.StatusBadRequest, err)
		return account, false
	}

	return account, true
}
