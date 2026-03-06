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

var nextblockwallets = []solana.PublicKey{
	solana.MustPublicKeyFromBase58("NexTbLoCkWykbLuB1NkjXgFWkX9oAtcoagQegygXXA2"),
	solana.MustPublicKeyFromBase58("NeXTBLoCKs9F1y5PJS9CKrFNNLU1keHW71rfh7KgA1X"),
	solana.MustPublicKeyFromBase58("NexTBLockJYZ7QD7p2byrUa6df8ndV2WSd8GkbWqfbb"),
	solana.MustPublicKeyFromBase58("neXtBLock1LeC67jYd1QdAa32kbVeubsfPNTJC1V5At"),
	solana.MustPublicKeyFromBase58("nEXTBLockYgngeRmRrjDV31mGSekVPqZoMGhQEZtPVG"),
	solana.MustPublicKeyFromBase58("NEXTbLoCkB51HpLBLojQfpyVAMorm3zzKg7w9NFdqid"),
	solana.MustPublicKeyFromBase58("nextBLoCkPMgmG8ZgJtABeScP35qLa2AMCNKntAP7Xc"),
	solana.MustPublicKeyFromBase58("NextbLoCkVtMGcV47JzewQdvBpLqT9TxQFozQkN98pE"),
}

type Payload struct {
	Transaction TransactionMessage `json:"transaction"`
}

type TransactionMessage struct {
	Content string `json:"content"`
}

type NextblockSend struct {
	apiKey string
	url    string
}

func NewNextblockSend(location string, api string) *NextblockSend {
	var url string
	if location == "frankfurt" {
		url = "https://fra.nextblock.io/api/v2/submit"
	} else if location == "newyork" {
		url = "https://ny.nextblock.io/api/v2/submit"
	} else {
		slog.Error("Invalid location, PLEASE ENTER A VALID NEXTBLOCK LOCATION")
		return nil
	}
	return &NextblockSend{
		apiKey: api,
		url:    url,
	}
}

func (n NextblockSend) SendTransaction(transaction solana.Transaction, _ SendParams) (success bool, err error) {
	cli := http.Client{}
	tx64Str := transaction.MustToBase64()
	txMsg := TransactionMessage{
		Content: tx64Str,
	}
	p := Payload{
		Transaction: txMsg,
	}
	t, _ := json.Marshal(p)

	req, err := http.NewRequest("POST", n.url, bytes.NewBuffer(t))
	if err != nil || req == nil {
		slog.Error(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("authorization", n.apiKey)

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

func (n NextblockSend) GetTipInstructions(sender solana.Wallet, TipFee float64) (instructions []solana.Instruction, err error) {
	wallet := nextblockwallets[rand.Intn(len(nextblockwallets))]
	transferIx := system.NewTransferInstruction(uint64(TipFee*1e9), sender.PublicKey(), wallet).Build()
	instructions = append(instructions, transferIx)
	return instructions, nil
}
