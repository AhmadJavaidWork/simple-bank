package gapi

import (
	"context"
	"math"

	db "github.com/ahmadjavaidwork/simple-bank/db/sqlc"
	"github.com/ahmadjavaidwork/simple-bank/pb"
	"github.com/ahmadjavaidwork/simple-bank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) ListAccount(
	ctx context.Context,
	req *pb.ListAccountRequest,
) (*pb.ListAccountResponse, error) {
	authPayload, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	violations := validateListAccountRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	arg := db.ListAccountsParams{
		Owner:  authPayload.Username,
		Limit:  req.PageSize,
		Offset: (req.PageId - 1) * req.PageSize,
	}

	accounts, err := server.store.ListAccounts(ctx, arg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal server error: %s", err)
	}

	rsp := &pb.ListAccountResponse{
		Account: []*pb.Account{},
	}
	for _, account := range accounts {
		rsp.Account = append(rsp.Account, convertAccount(account))
	}

	return rsp, nil
}

func validateListAccountRequest(
	req *pb.ListAccountRequest,
) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateInt(int(req.GetPageId()), 1, math.MaxInt32); err != nil {
		violations = append(violations, fieldViolation("page_size", err))
	}
	if err := val.ValidateInt(int(req.GetPageSize()), 5, 10); err != nil {
		violations = append(violations, fieldViolation("page_id", err))
	}
	return violations
}
