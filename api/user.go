package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	db "github.com/billiraheem/Billi-Bank/db/sqlc"
	"github.com/billiraheem/Billi-Bank/token"
	"github.com/billiraheem/Billi-Bank/utils"
	"github.com/billiraheem/Billi-Bank/worker"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/lib/pq"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	Fullname string `json:"fullname" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type userResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"fullname"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		Username:          user.Username,
		FullName:          user.Fullname,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
}

func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errRes(err))
		return
	}

	hasedPasswword, err := utils.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errRes(err))
		return
	}

	args := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hasedPasswword,
		Fullname:       req.Fullname,
		Email:          req.Email,
	}

	user, err := server.store.CreateUser(ctx, args)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errRes(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errRes(err))
		return
	}

	// TODO: use db transaction 
	// send verify email to user
	taskPayload := &worker.PayloadSendVerifyEmail{
		Username: user.Username,
	}
	opts := []asynq.Option{
		asynq.MaxRetry(10),
		asynq.ProcessIn(10 * time.Second),
		asynq.Queue(worker.QueueCritical),
	}
	err = server.taskDistributor.DistributeTaskSendVerifyEmail(ctx, taskPayload, opts...)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errRes(err))
		return
	}

	res := newUserResponse(user)

	ctx.JSON(http.StatusOK, res)
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	SessionID             uuid.UUID    `json:"session_id"`
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  userResponse `json:"user"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errRes(err))
		return
	}

	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errRes(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errRes(err))
		return
	}

	err = utils.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errRes(err))
		return
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(req.Username, server.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errRes(err))
		return
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(req.Username, server.config.RefreshTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errRes(err))
		return
	}

	session, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    ctx.Request.UserAgent(),
		ClientIp:     ctx.ClientIP(),
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})

	res := loginUserResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  newUserResponse(user),
	}

	ctx.JSON(http.StatusOK, res)
}

type updateUserRequest struct {
	Password *string `json:"password" binding:"omitempty,min=6"`
	Fullname *string `json:"fullname" binding:"omitempty"`
	Email    *string `json:"email" binding:"omitempty,email"`
}

func (server *Server) updateUser(ctx *gin.Context) {
	var req updateUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errRes(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayload).(*token.Payload)

	args := db.UpdateUserParams{
		Username: authPayload.Username,
		Fullname: sql.NullString{
			String: utils.DerefString(req.Fullname),
			Valid:  req.Fullname != nil,
		},
		Email: sql.NullString{
			String: utils.DerefString(req.Email),
			Valid:  req.Email != nil,
		},
	}

	if req.Password != nil {
		hasedPasswword, err := utils.HashPassword(*req.Password)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errRes(err))
			return
		}

		args.HashedPassword = sql.NullString{
			String: hasedPasswword,
			Valid:  req.Password != nil,
		}

		args.PasswordChangedAt = sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}
	}

	updatedUser, err := server.store.UpdateUser(ctx, args)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errRes(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errRes(err))
		return
	}

	res := newUserResponse(updatedUser)

	ctx.JSON(http.StatusOK, res)
}

type logoutUserRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (server *Server) logoutUser(ctx *gin.Context) {
	var req logoutUserRequest

    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, errRes(err))
        return
    }

    // verify the refresh token is valid
    refreshPayload, err := server.tokenMaker.VerifyToken(req.RefreshToken)
    if err != nil {
		fmt.Println("VERIFY REFRESH TOKEN ERROR:", err)

        ctx.JSON(http.StatusUnauthorized, errRes(err))
        return
	}

	// get the session
    session, err := server.store.GetSession(ctx, refreshPayload.ID)
    if err != nil {
        if err == sql.ErrNoRows {
            ctx.JSON(http.StatusNotFound, errRes(err))
            return
        }
        ctx.JSON(http.StatusInternalServerError, errRes(err))
        return
    }

	// make sure session belongs to the right user
    authPayload := ctx.MustGet(authorizationPayload).(*token.Payload)
	fmt.Println("SESSION USERNAME:", session.Username)
	fmt.Println("AUTH USERNAME:", authPayload.Username)
    if session.Username != authPayload.Username {
		fmt.Println("SESSION USERNAME DOES NOT MATCH AUTH USERNAME")
        err := errors.New("session does not belong to the authenticated user")
        ctx.JSON(http.StatusUnauthorized, errRes(err))
        return
    }

	// block the session
    _, err = server.store.BlockSession(ctx, session.ID)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, errRes(err))
        return
    }

    ctx.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}