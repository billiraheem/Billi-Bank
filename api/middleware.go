package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/billiraheem/Billi-Bank/token"
	"github.com/gin-gonic/gin"
)

const (
	authorizationHeaderKey = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayload = "authorization_payload"
)

func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fmt.Println("AUTH MIDDLEWARE HIT")

		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
		fmt.Println("Authorization header:", authorizationHeader)
		if len(authorizationHeader) == 0 {
			err := errors.New("authorization header is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errRes(err))
			return 
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			err := errors.New("invalid authorization header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errRes(err))
			return 
		}

		authorizationType := strings.ToLower(fields[0])
		fmt.Println("Authorization type:", authorizationType)
		if authorizationType != authorizationTypeBearer {
			err := fmt.Errorf("unsupported authorization type %s", authorizationType)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errRes(err))
			return 
		}

		accessToken := fields[1]
		fmt.Println("Access token received:", accessToken)
		payload, err :=tokenMaker.VerifyToken(accessToken)
		if err != nil {
			fmt.Println("VERIFY ACCESS TOKEN ERROR:", err)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errRes(err))
			return
		}

		fmt.Println("ACCESS TOKEN VALID")
		fmt.Println("Username:", payload.Username)
		fmt.Println("Session ID:", payload.ID)


		ctx.Set(authorizationPayload, payload)
		ctx.Next()
	}
}