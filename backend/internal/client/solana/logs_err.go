package solana

import (
	"strings"

	"dd-prediction-api/pkg/apperrors"

	"go.uber.org/zap"
)

var (
	ErrInsufficientFunds = apperrors.PaymentRequired("insufficient funds for proceeding a transaction")
)

func ParseLogsForError(logs []string) error {
	for _, log := range logs {
		switch {
		case IsInsufficientFundForCommission(log):
			return ErrInsufficientFunds
		}
	}
	zap.L().Error("solana error logs", zap.Strings("logs", logs))
	return apperrors.Internal("transaction: result err")
}

func IsInsufficientFundForCommission(log string) bool {
	return strings.Contains(log, "insufficient")
}
