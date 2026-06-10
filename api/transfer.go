package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	db "github.com/billiraheem/Billi-Bank/db/sqlc"
	"github.com/billiraheem/Billi-Bank/token"
	"github.com/gin-gonic/gin"
)

type transferRequest struct {
	FromAccountID int64 `json:"from_account_id" binding:"required,min=1"`
	ToAccountID int64 `json:"to_account_id" binding:"required,min=1"`
	Amount int64 `json:"amount" binding:"required,gt=0"`
	Currency string `json:"currency" binding:"required,currency"`
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req transferRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errRes(err))
		return
	}

	fromAccount, valid := server.validAccount(ctx, req.FromAccountID, req.Currency)
	if !valid {
		return
	}

	authPayload := ctx.MustGet(authorizationPayload).(*token.Payload)
	if fromAccount.Owner != authPayload.Username {
		err := errors.New("account does not belong to user")
		ctx.JSON(http.StatusUnauthorized, errRes(err))
		return
	}

	_, valid = server.validAccount(ctx, req.ToAccountID, req.Currency)
	if !valid {
		return
	}

	args := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID: req.ToAccountID,
		Amount: req.Amount,
	}

	result, err := server.store.TransferTx(ctx, args)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errRes(err))
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (server *Server) validAccount(ctx *gin.Context, accountID int64, currency string) (db.Account, bool) {
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errRes(err))
			return account, false
		}

		ctx.JSON(http.StatusInternalServerError, errRes(err))
		return account, false
	}

	if account.Currency != currency {
		err := fmt.Errorf("account [%d] currency mismatch %s vs %s", account.ID, account.Currency, currency)
		ctx.JSON(http.StatusBadRequest, errRes(err))
		return account, false
	}

	return account, true
}