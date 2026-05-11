package sigtracker

import (
	"context"

	solanalib "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

type rpcClientInterface interface {
	GetSignatureStatuses(ctx context.Context, searchTransactionHistory bool, signatures ...solanalib.Signature) (*rpc.GetSignatureStatusesResult, error)
}

type signatureSubscriptionInterface interface {
	Recv(ctx context.Context) (*ws.SignatureResult, error)
	Unsubscribe()
}

type wsClientInterface interface {
	SignatureSubscribe(signature solanalib.Signature, commitment rpc.CommitmentType) (signatureSubscriptionInterface, error)
	Close()
}

type rpcClientAdapter struct {
	real *rpc.Client
}

func (a rpcClientAdapter) GetSignatureStatuses(ctx context.Context, search bool, signatures ...solanalib.Signature) (*rpc.GetSignatureStatusesResult, error) {
	return a.real.GetSignatureStatuses(ctx, search, signatures...)
}

type wsClientAdapter struct {
	real *ws.Client
}

func (a wsClientAdapter) SignatureSubscribe(signature solanalib.Signature, c rpc.CommitmentType) (signatureSubscriptionInterface, error) {
	sub, err := a.real.SignatureSubscribe(signature, c)
	if err != nil {
		return nil, err
	}
	return subAdapter{sub}, nil
}

func (a wsClientAdapter) Close() {
	a.real.Close()
}

type subAdapter struct {
	real *ws.SignatureSubscription
}

func (s subAdapter) Recv(ctx context.Context) (*ws.SignatureResult, error) {
	return s.real.Recv(ctx)
}
func (s subAdapter) Unsubscribe() {
	s.real.Unsubscribe()
}
