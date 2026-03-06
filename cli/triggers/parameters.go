package triggers

import "github.com/gagliardetto/solana-go/rpc"

type RPCOpts struct {
	Client  *rpc.Client
	Retries int
	DelayMS int64
}
