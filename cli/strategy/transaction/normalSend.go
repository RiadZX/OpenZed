package strategy

import (
	"Zed/strategy"
	"context"
	"errors"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gookit/slog"
	"time"
)

type NormalSend struct {
	tps        int
	httpClient IClient
}

func NewNormalSend(tps int, httpClient IClient) *NormalSend {
	return &NormalSend{
		tps:        tps,
		httpClient: httpClient,
	}
}

func (n NormalSend) SendTransaction(tx solana.Transaction, params SendParams) (success bool, err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Add the signature to the buffer
	strategy.SigBuffer.Append(tx.Signatures[0].String())
	timeOutTime := time.Now().Add(20 * time.Second)
	txChan := make(chan struct{}, n.tps)
	slog.Info("Sending transactions")

	for {
		timeStart := time.Now()
		if time.Now().After(timeOutTime) {
			break
		}

		txChan <- struct{}{}
		go func(ctx context.Context) {
			defer func() { <-txChan }()

			sig, err := sendSingleTransaction(n.httpClient, &tx)
			if err != nil {
				slog.Debug("Error sending transaction: ", err)
				return //continue sending
			}
			slog.Info(params.TaskName + ": tx sent : " + sig)
			status := strategy.SigBuffer.GetStatus(tx.Signatures[0].String())
			if status != strategy.TransactionPending {
				slog.Info(params.TaskName + ": tx sent : " + sig)
				cancel()
				return // exit from sender
			}
		}(ctx)

		select {
		case <-ctx.Done():
			if strategy.SigBuffer.GetStatus(tx.Signatures[0].String()) == strategy.TransactionSuccess {
				return true, nil
			}
			if strategy.SigBuffer.GetStatus(tx.Signatures[0].String()) == strategy.TransactionError {
				return false, errors.New("transaction failed")
			}
			break
		default:
			timeEnd := time.Now()
			timeTaken := timeEnd.Sub(timeStart)
			timeToSleep := time.Second / time.Duration(n.tps)
			if timeTaken < timeToSleep {
				time.Sleep(timeToSleep - timeTaken)
			}
		}
	}

	return false, errors.New("timeout")
}

// SendTransaction sends the transaction in the most optimal way
// simulationFailed will force the transaction to be sent normally, not using JITO
func sendSingleTransaction(client IClient, tx *solana.Transaction) (string, error) {
	if tx == nil {
		return "", errors.New("tx is nil")
	}
	sig, err := client.SendTransactionWithOpts(context.Background(), tx, rpc.TransactionOpts{SkipPreflight: true, MaxRetries: new(uint)})
	if err != nil {
		slog.Debug("Normal sending failed: " + err.Error())
		return sig.String(), err
	}
	return sig.String(), nil
}

func (n NormalSend) GetTipInstructions(sender solana.Wallet, TipFee float64) (instructions []solana.Instruction, err error) {
	// no tip for normal send
	return []solana.Instruction{}, nil
}
