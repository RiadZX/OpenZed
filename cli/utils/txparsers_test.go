package utils

import (
	"testing"
)

func Test_GetPostTokenBalance(t *testing.T) {
	//client := rpc.New("https://api.mainnet-beta.solana.com")
	//sig := solana.MustSignatureFromBase58("3VsCjNbk1Bk7wmbwL4RQwmNqBXforsdeLq3WLqtg7gijTgKYhnMLzLqppcDPxx49RTpBK7VZ8Ki1CP4xxzongBYx")
	//var maxSupportedTransactionVersion uint64 = 1
	//tx, err := client.GetTransaction(context.Background(), sig, &rpc.GetTransactionOpts{MaxSupportedTransactionVersion: &maxSupportedTransactionVersion})
	//if err != nil {
	//	return
	//}
	//balances := CustomPrePostTokenBalances{PreTokenBalances: make([]XTokenBalance, 0), PostTokenBalances: make([]XTokenBalance, 0)}
	//err = balances.FromSolanaPrePro(tx.Meta.PreTokenBalances, tx.Meta.PostTokenBalances)
	//if err != nil {
	//	return
	//}
	//
	//owner := "BuqZaKUiwUSrcLDMPbUNSkRZBXLxFZZieHQVf4qgveLo"
	//mint := "FzAwkijFzSga76USEpo3126GyTRmxcMDbA9TYD2PPBt5"
	//balance := balances.GetPostTokenBalance(owner, mint)
	//balance2 := balances.GetPreTokenBalance(owner, mint)
	//if balance == 0 && balance2 == 0 {
	//	t.Error("GetTokenBalance failed")
	//}
	//slog.Info("Sold: " + strconv.Itoa(int(balance2-balance)))
}
