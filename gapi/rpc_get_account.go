package gapi

import (
	"context"
	"database/sql"
	"math"
	"net/http"

	"github.com/ahmadjavaidwork/simple-bank/pb"
	"github.com/ahmadjavaidwork/simple-bank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) GetAccount(
	ctx context.Context,
	req *pb.GetAccountRequest,
) (*pb.GetAccountResponse, error) {
	_, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	violations := validateGetAccountRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	account, err := server.store.GetAccount(ctx, req.GetId())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "account not found: %s", err)
		}
		return nil, status.Errorf(http.StatusInternalServerError, "internal server error: %s", err)
	}

	rsp := &pb.GetAccountResponse{
		Account: convertAccount(account),
	}
	return rsp, nil
}

func validateGetAccountRequest(
	req *pb.GetAccountRequest,
) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateInt(req.GetId(), 1, math.MaxInt64); err != nil {
		violations = append(violations, fieldViolation("id", err))
	}
	return violations
}
