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

var bloxWallets = []solana.PublicKey{
	solana.MustPublicKeyFromBase58("95cfoy472fcQHaw4tPGBTKpn6ZQnfEPfBgDQx6gcRmRg"),
	solana.MustPublicKeyFromBase58("HWEoBxYs7ssKuudEjzjmpfJVX7Dvi7wescFsVx2L5yoY"),
}

type BloxRouteSend struct {
	apiKey string
	url    string
}

func NewBloxRouteSend(apiKey string, url string) *BloxRouteSend {
	//if it doesnt end with /api/v2/submit, return nil and log error
	if url[len(url)-14:] != "/api/v2/submit" {
		slog.Error("Invalid URL, make sure it ends with /api/v2/submit")
		return nil
	}
	return &BloxRouteSend{
		apiKey: apiKey,
		url:    url,
	}
}

type BloxPayload struct {
	Transaction            TransactionMessage `json:"transaction"`
	FrontRunningProtection bool               `json:"frontRunningProtection"`
	UseStakedRPCs          bool               `json:"useStakedRPCs"`
}

func (br BloxRouteSend) SendTransaction(transaction solana.Transaction, _ SendParams) (success bool, err error) {
	cli := http.Client{}
	tx64Str := transaction.MustToBase64()
	txMsg := TransactionMessage{
		Content: tx64Str,
	}
	p := BloxPayload{
		Transaction:            txMsg,
		FrontRunningProtection: false,
		UseStakedRPCs:          true,
	}
	t, _ := json.Marshal(p)

	req, err := http.NewRequest("POST", br.url, bytes.NewBuffer(t))
	if err != nil || req == nil {
		slog.Error(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", br.apiKey)

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

func (br BloxRouteSend) GetTipInstructions(sender solana.Wallet, TipFee float64) (instructions []solana.Instruction, err error) {
	wallet := bloxWallets[rand.Intn(len(bloxWallets))]
	transferIx := system.NewTransferInstruction(uint64(TipFee*1e9), sender.PublicKey(), wallet).Build()
	instructions = append(instructions, transferIx)
	return instructions, nil
}
