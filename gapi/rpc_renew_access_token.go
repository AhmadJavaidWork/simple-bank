package gapi

import (
	"context"
	"database/sql"
	"time"

	"github.com/ahmadjavaidwork/simple-bank/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server) RenewAccessToken(
	ctx context.Context,
	req *pb.RenewAccessTokenRequest,
) (*pb.RenewAccessTokenResponse, error) {
	refreshPayload, err := server.tokenMaker.VerifyToken(req.GetRefreshToken())
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "cannot verify token: %s", err)
	}

	session, err := server.store.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token")
		}
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	if session.IsBlocked {
		return nil, status.Errorf(codes.Unauthenticated, "blocked session")
	}

	if session.Username != refreshPayload.Username {
		return nil, status.Errorf(codes.Unauthenticated, "incorrect session user")
	}

	if session.RefreshToken != req.RefreshToken {
		return nil, status.Errorf(codes.Unauthenticated, "mismatch session token")
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, status.Errorf(codes.Unauthenticated, "expired session")
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		refreshPayload.Username,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to create access token: %s", err)
	}

	rsp := &pb.RenewAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: timestamppb.New(accessPayload.ExpiredAt),
	}

	return rsp, nil
}
