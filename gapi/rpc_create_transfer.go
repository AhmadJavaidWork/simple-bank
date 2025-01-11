package gapi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"

	db "github.com/ahmadjavaidwork/simple-bank/db/sqlc"
	"github.com/ahmadjavaidwork/simple-bank/pb"
	"github.com/ahmadjavaidwork/simple-bank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateTransfer(
	ctx context.Context,
	req *pb.CreateTransferRequest,
) (*pb.CreateTransferResponse, error) {
	authPayload, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	violations := validateCreateTransferRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	fromAccount, statusCode, err := server.validAccount(
		ctx,
		req.GetFromAccountId(),
		req.GetCurrency(),
	)
	if err != nil {
		return nil, status.Errorf(statusCode, "from %s", err)
	}

	if fromAccount.Owner != authPayload.Username {
		return nil, status.Errorf(codes.PermissionDenied, "from account doesn't belong to the authenticated user")
	}

	_, statusCode, err = server.validAccount(
		ctx,
		req.GetToAccountId(),
		req.GetCurrency(),
	)
	if err != nil {
		return nil, status.Errorf(statusCode, "to %s", err)
	}

	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountId,
		ToAccountID:   req.ToAccountId,
		Amount:        req.Amount,
	}

	result, err := server.store.TransferTx(ctx, arg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal server error: %s", err)
	}

	rsp := convertTransferResponse(result)

	return rsp, nil
}

func (server *Server) validAccount(
	ctx context.Context,
	accountID int64,
	currency string,
) (db.Account, codes.Code, error) {
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return account, codes.NotFound, fmt.Errorf("account with id: %d not found", accountID)
		}
		return account, codes.Internal, err
	}

	if account.Currency != currency {
		err := fmt.Errorf("account [%d] currency mismatch: %s vs %s", accountID, account.Currency, currency)
		return account, codes.InvalidArgument, err
	}

	return account, codes.OK, nil
}

func validateCreateTransferRequest(
	req *pb.CreateTransferRequest,
) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateInt(req.GetFromAccountId(), 1, math.MaxInt64); err != nil {
		violations = append(violations, fieldViolation("from_account_id", err))
	}
	if err := val.ValidateInt(req.GetToAccountId(), 1, math.MaxInt64); err != nil {
		violations = append(violations, fieldViolation("to_account_id", err))
	}
	if err := val.ValidateInt(req.GetAmount(), 1, math.MaxInt64); err != nil {
		violations = append(violations, fieldViolation("amount", err))
	}
	if err := val.ValidateCurrency(req.GetCurrency()); err != nil {
		violations = append(violations, fieldViolation("currency", err))
	}
	return violations
}
