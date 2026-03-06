package strategy

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gookit/slog"
	"io"
	"math/rand"
	"net/http"
)

var circularwallets = []solana.PublicKey{
	solana.MustPublicKeyFromBase58("FAST3dMFZvESiEipBvLSiXq3QCV51o3xuoHScqRU6cB6"),
	solana.MustPublicKeyFromBase58("FASTHPW6akdGh9PFSdhMTbCuGkCSX7LsUjjnaB2RTQ4v"),
	solana.MustPublicKeyFromBase58("FASTYKWXRfAoty7SQCM1mGVrmPUyyNcF4tc3DUkLDAu9"),
	solana.MustPublicKeyFromBase58("FASTPB76TxKPMZ7Q29m8v4zJn8gUjbWyvTEQaaxhwN7M"),
	solana.MustPublicKeyFromBase58("FASTs6ctgbsuZegMzUs4DPUYhRSZUPCjgCVnttHbpQAp"),
	solana.MustPublicKeyFromBase58("FASTYmSidNfLwdwiQEhCTtzghxEtaipeNSDSwh9xDPs3"),
	solana.MustPublicKeyFromBase58("FASTCKnwwY6iL3CknRgg3Zqir7jeagDDhxSnBQQy5a1C"),
	solana.MustPublicKeyFromBase58("FASTKL1AamNKrwnvbKwo4PU8434BBdqVrTtugM6oDU71"),
}

type CircularSend struct {
	apiKey string
}

func NewCircularSend(apiKey string) *CircularSend {
	return &CircularSend{
		apiKey: apiKey,
	}
}

type circularBody struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type circularParams struct {
	Options struct {
		FrontRunningProtection bool `json:"frontRunningProtection"`
	}
}

func (n CircularSend) SendTransaction(transaction solana.Transaction, _ SendParams) (success bool, err error) {
	cli := http.Client{}
	tx64Str := transaction.MustToBase64()

	body := circularBody{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "sendTransaction",
		Params: []interface{}{
			tx64Str,
			circularParams{
				Options: struct {
					FrontRunningProtection bool `json:"frontRunningProtection"`
				}{FrontRunningProtection: false},
			},
		},
	}

	t, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", "https://fast.circular.bot/transactions", bytes.NewBuffer(t))
	if err != nil || req == nil {
		slog.Error(err)
		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-api-key", n.apiKey)

	do, err := cli.Do(req)
	if err != nil || do == nil {
		slog.Error(err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			slog.Error(err)
		}
	}(do.Body)
	b, _ := io.ReadAll(do.Body)
	slog.Info(string(b))
	if do.StatusCode == 200 {
		slog.Info("Sent transaction with signature ", transaction.Signatures[0].String())
		return true, nil
	}
	return false, errors.New(string(b))
}

func (n CircularSend) GetTipInstructions(sender solana.Wallet, TipFee float64) (instructions []solana.Instruction, err error) {
	wallet := circularwallets[rand.Intn(len(circularwallets))]
	transferIx := system.NewTransferInstruction(uint64(TipFee*1e9), sender.PublicKey(), wallet).Build()
	instructions = append(instructions, transferIx)
	return instructions, nil
}
