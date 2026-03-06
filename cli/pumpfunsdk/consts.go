package pumpfunsdk

import "github.com/gagliardetto/solana-go"

var (
	PUMP_PROGRAM_ADDRESS = solana.MustPublicKeyFromBase58("6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P")
	GLOBAL_ACCOUNT       = solana.MustPublicKeyFromBase58("4wTV1YmiEkRvAtNtsSGPtUrqRYQMe5SKy2uB4Jjaxnjf")
	FEE_RECIPIENT        = solana.MustPublicKeyFromBase58("62qc2CNXwrYqQScmEdiZFFAnJR262PxWEuNQtxfafNgV")
	EVENT_AUTHORITY      = solana.MustPublicKeyFromBase58("Ce6TQqeHC9p8KetsN6JsjHK7UTZk7nasjjnr7XxXp9F1")
	ATA_TOKEN_PROGRAM    = solana.MustPublicKeyFromBase58("ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL")
	PUMP_AMM             = solana.MustPublicKeyFromBase58("pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA")
	AMM_FEE_RECIPIENTS   = []solana.PublicKey{
		solana.MustPublicKeyFromBase58("7hTckgnGnLQR6sdH7YkqFTAA7VwTfYFaZ6EhEsU3saCX"),
		solana.MustPublicKeyFromBase58("62qc2CNXwrYqQScmEdiZFFAnJR262PxWEuNQtxfafNgV"),
		solana.MustPublicKeyFromBase58("JCRGumoE9Qi5BBgULTgdgTLjSgkCMSbF62ZZfGs84JeU"),
		solana.MustPublicKeyFromBase58("FWsW1xNtWscwNmKv6wVsU1iTzRN6wmmk3MjxRP5tT7hz"),
	}
)

type TransactionType int

const (
	Buy        TransactionType = iota // The transaction is a buy transaction. Token bought.
	Sell                              // The transaction is a sell transaction. Token sold.
	AMMBuy                            // The transaction is an AMMBuy transaction.
	AMMSell                           // The transaction is an AMMSell transaction.
	Create                            // The transaction is a creation transaction. New token created.
	BuySell                           // The transaction is a buy and sell transaction. Token bought and sold.
	CreateBuy                         // The transaction is a create and buy transaction. New token created and bought.
	Error                             // The transaction is an error transaction. This has failed. Or error during parsing.
	NotPumpFun                        // The transaction is not a pumpfun transaction.
	Migration                         // The transaction is a migration transaction. Token migrated.
)
