package gapi

import (
	"context"
	"database/sql"

	db "github.com/billiraheem/Billi-Bank/db/sqlc"
	"github.com/billiraheem/Billi-Bank/pb"
	"github.com/billiraheem/Billi-Bank/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	user, err := server.store.GetUser(ctx, req.GetUsername())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user does not exist: %s", err)
		}

		return nil, status.Errorf(codes.Internal, "failed to fetch user: %s", err)
	}

	err = utils.CheckPassword(req.GetPassword(), user.HashedPassword)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "incorrect password: %s", err)
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(req.Username, server.config.AccessTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create access token: %s", err)
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(req.Username, server.config.RefreshTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create refresh token: %s", err)
	}

	mtdt := server.extractMetadata(ctx)
	session, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    mtdt.UserAgent,
		ClientIp:     mtdt.ClientIP,
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})

	rsp := &pb.LoginUserResponse{
		SessionId: session.ID.String(),
		AccessToken: accessToken,
		AccessTokenExpiresAt: timestamppb.New(accessPayload.ExpiredAt),
		RefreshToken: refreshToken,
		RefreshTokenExpiresAt: timestamppb.New(refreshPayload.ExpiredAt),
		User: convertUser(user),
	}

	return rsp, nil
}

// i will say there is no need to return the actual err to the client but leaving it since this is a personal project