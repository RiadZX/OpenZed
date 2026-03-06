package models

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type IClient interface {
	GetTokenAccountsByOwner(ctx context.Context, owner solana.PublicKey, conf *rpc.GetTokenAccountsConfig, opts *rpc.GetTokenAccountsOpts) (out *rpc.GetTokenAccountsResult, err error)
	GetTransaction(ctx context.Context, txSig solana.Signature, opts *rpc.GetTransactionOpts) (out *rpc.GetTransactionResult, err error)
	GetAccountInfoWithOpts(ctx context.Context, account solana.PublicKey, opts *rpc.GetAccountInfoOpts) (*rpc.GetAccountInfoResult, error)
	SendTransactionWithOpts(ctx context.Context, transaction *solana.Transaction, opts rpc.TransactionOpts) (signature solana.Signature, err error)
	GetLatestBlockhash(ctx context.Context, commitment rpc.CommitmentType) (out *rpc.GetLatestBlockhashResult, err error)
	SimulateTransactionWithOpts(ctx context.Context, transaction *solana.Transaction, opts *rpc.SimulateTransactionOpts) (out *rpc.SimulateTransactionResponse, err error)
	GetSignatureStatuses(ctx context.Context, searchTransactionHistory bool, transactionSignatures ...solana.Signature) (out *rpc.GetSignatureStatusesResult, err error)
	GetTokenAccountBalance(ctx context.Context, account solana.PublicKey, commitment rpc.CommitmentType) (out *rpc.GetTokenAccountBalanceResult, err error)
}
