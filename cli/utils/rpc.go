package utils

import (
	"Zed/models"
	"context"
	"errors"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"math"
	"math/rand"
	"strconv"
	"time"
)

var (
	maxSupportedTransactionVersion uint64 = 1
)

func GetTokenBalance(client models.IClient, account solana.PublicKey, delay int64, retries int64) (uiTokenAmount rpc.UiTokenAmount, err error) {
	if retries == 0 {
		return rpc.UiTokenAmount{}, errors.New("retries exhausted for adress: " + account.String())
	}
	if client == nil {
		return rpc.UiTokenAmount{}, errors.New("client or address is nil")
	}
	tokenBalance, err := client.GetTokenAccountBalance(context.TODO(), account, rpc.CommitmentProcessed)
	if err != nil {
		time.Sleep(time.Duration(delay) * time.Millisecond)
		return GetTokenBalance(client, account, delay, retries-1)
	}
	if tokenBalance == nil {
		return rpc.UiTokenAmount{}, errors.New("token balance not found")
	}
	if tokenBalance.Value == nil {
		return rpc.UiTokenAmount{}, errors.New("token balance value not found")
	}
	_, err = strconv.Atoi(tokenBalance.Value.Amount)
	if err != nil {
		return rpc.UiTokenAmount{}, err
	}
	return *tokenBalance.Value, nil
}
func exponentialBackoff(attempt int, baseDelay time.Duration, maxDelay time.Duration) time.Duration {
	delay := baseDelay * time.Duration(math.Pow(2, float64(attempt)))
	if delay > maxDelay {
		return maxDelay
	}
	return delay
}
func GetTransaction(ctx context.Context, client models.IClient, signature string, rpcs []string, baseDelayMS int64, maxRetries int64) (*rpc.GetTransactionResult, error) {
	var err error
	var out *rpc.GetTransactionResult

	baseDelay := time.Duration(baseDelayMS) * time.Millisecond
	maxDelay := time.Second * 5 // Adjust max delay as needed
	for attempt := int64(0); attempt < maxRetries; attempt++ {
		if len(rpcs) != 0 {
			url := rpcs[rand.Intn(len(rpcs))]
			client = rpc.New(url)
		}
		//else just use the client
		out, err = client.GetTransaction(ctx, solana.MustSignatureFromBase58(signature),
			&rpc.GetTransactionOpts{
				Encoding:                       solana.EncodingBase64,
				Commitment:                     rpc.CommitmentConfirmed,
				MaxSupportedTransactionVersion: &maxSupportedTransactionVersion,
			})
		if err == nil {
			return out, nil
		}

		delay := exponentialBackoff(int(attempt), baseDelay, maxDelay)
		time.Sleep(delay)
	}

	return nil, errors.New("retries exhausted for signature: " + signature)
}
