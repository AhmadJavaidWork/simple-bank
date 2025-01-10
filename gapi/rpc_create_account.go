package gapi

import (
	"context"

	db "github.com/ahmadjavaidwork/simple-bank/db/sqlc"
	"github.com/ahmadjavaidwork/simple-bank/pb"
	"github.com/ahmadjavaidwork/simple-bank/val"
	"github.com/lib/pq"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateAccount(
	ctx context.Context,
	req *pb.CreateAccountRequest,
) (*pb.CreateAccountResponse, error) {
	authPayload, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	violations := validateCreateAccountRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	arg := db.CreateAccountParams{
		Owner:    authPayload.Username,
		Currency: req.GetCurrency(),
		Balance:  0,
	}

	account, err := server.store.CreateAccount(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				return nil, status.Errorf(codes.PermissionDenied, "cannot create account: %s", err)
			}
			return nil, status.Errorf(codes.Internal, "cannot create account: %s", err)
		}
	}

	rsp := &pb.CreateAccountResponse{
		Account: convertAccount(account),
	}

	return rsp, nil
}

func validateCreateAccountRequest(
	req *pb.CreateAccountRequest,
) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateCurrency(req.GetCurrency()); err != nil {
		violations = append(violations, fieldViolation("currency", err))
	}

	return violations
}
