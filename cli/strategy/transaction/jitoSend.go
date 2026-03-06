package strategy

import (
	"Zed/strategy"
	"errors"
	"github.com/gagliardetto/solana-go"
	"github.com/gookit/slog"
	"github.com/weeaa/jito-go/clients/searcher_client"
	"strings"
	"time"
)

type JitoSend struct {
	httpClient IClient
	jitoClient *searcher_client.Client
	tps        int
}

func NewJitoSend(httpClient IClient, jitoClient *searcher_client.Client, tps int) *JitoSend {
	return &JitoSend{
		jitoClient: jitoClient,
		httpClient: httpClient,
		tps:        tps,
	}
}

func (n JitoSend) SendTransaction(tx solana.Transaction, params SendParams) (success bool, error error) {
	strategy.SigBuffer.Append(tx.Signatures[0].String())
	iterations := 20 * n.tps // Calculate the number of iterations based on tps
	for i := 0; i < iterations; i++ {
		slog.Info("Sending transaction")
		sig, err := n.sendJitoTransaction(&tx)
		if err != nil && strings.Contains(err.Error(), "bundle contains an already processed transaction") {
			strategy.SigBuffer.SetStatusAndLog(tx.Signatures[0].String(), strategy.TransactionSuccess)
			return true, nil
		}
		if strategy.SigBuffer.GetStatus(tx.Signatures[0].String()) == strategy.TransactionSuccess {
			return true, nil
		}
		if err != nil {
			if strings.Contains(err.Error(), "bundle contains an already processed transaction") {
				strategy.SigBuffer.SetStatusAndLog(tx.Signatures[0].String(), strategy.TransactionSuccess)
				return true, nil
			}
			slog.Debug("Error sending transaction: ", err)
			time.Sleep(time.Duration(1000/n.tps) * time.Millisecond)
			continue
		}
		slog.Info(params.TaskName + ": Bundle sent : " + sig)
		time.Sleep(time.Duration(1000/n.tps) * time.Millisecond)
	}
	return false, nil
}

func (n JitoSend) sendJitoTransaction(tx *solana.Transaction) (string, error) {
	if tx == nil {
		return "", errors.New("tx is nil")
	}
	resp, err := n.jitoClient.BroadcastBundle([]*solana.Transaction{tx})
	if err != nil {
		if strings.Contains(err.Error(), "bundle contains an already processed transaction") {
			if resp != nil {
				return resp.Uuid, nil
			} else {
				return "---", err
			}
		}
		slog.Debug("Jito sending failed: " + err.Error())
		return "err sending", err
	}
	if resp == nil {
		return "---2", nil
	}
	return resp.Uuid, nil
}

func (n JitoSend) GetTipInstructions(sender solana.Wallet, TipFee float64) (instructions []solana.Instruction, err error) {
	tipInst, err := n.jitoClient.GenerateTipRandomAccountInstruction(uint64(TipFee*1e9), sender.PublicKey())
	if err != nil {
		slog.Debug("Error generating jito tip instruction: " + err.Error())
		return nil, err
	}
	instructions = append(instructions, tipInst)
	return instructions, nil
}
