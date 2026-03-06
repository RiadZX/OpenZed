package utils

import (
	"Zed/models"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/program/associated_token_account"
	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gookit/slog"
	"io/ioutil"
	"net/http"
)

func SignatureLink(sig string) string {
	return fmt.Sprintf("https://solana.fm/tx/%s", sig)
}

func GetPriorityFeeEstimate(url string, level string, program string) (uint64, error) {
	acc := ""
	if program == "PumpFun" {
		acc = "6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P"
	}
	if program == "Raydium" {
		acc = "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"
	}
	if acc == "" {
		return 0, errors.New("invalid program")
	}
	requestBody := RequestBodyPriorityFee{
		Jsonrpc: "2.0",
		ID:      "getPriorityFeeEstimate",
		Method:  "getPriorityFeeEstimate",
		Params: []struct {
			AccountKeys []string        `json:"accountKeys"`
			Options     map[string]bool `json:"options"`
		}{
			{
				AccountKeys: []string{acc},
				Options: map[string]bool{
					"includeAllPriorityFeeLevels": true,
				},
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return 0, fmt.Errorf("error marshaling request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return 0, fmt.Errorf("error sending request: %w, status code: %d", err, resp.StatusCode)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading response body: %w", err)
	}

	var apiResponse ApiResponse
	err = json.Unmarshal(respBody, &apiResponse)
	if err != nil {
		return 0, fmt.Errorf("error unmarshaling response: %w", err)
	}

	var feeLevel float64
	switch level {
	case "min":
		feeLevel = apiResponse.Result.PriorityFeeLevels.Min
	case "low":
		feeLevel = apiResponse.Result.PriorityFeeLevels.Low
	case "medium":
		feeLevel = apiResponse.Result.PriorityFeeLevels.Medium
	case "high":
		feeLevel = apiResponse.Result.PriorityFeeLevels.High
	case "veryHigh":
		feeLevel = apiResponse.Result.PriorityFeeLevels.VeryHigh
	case "unsafeMax":
		feeLevel = apiResponse.Result.PriorityFeeLevels.UnsafeMax
	default:
		return 0, errors.New("invalid priority fee level")
	}
	return uint64(feeLevel), nil
}

// GetUnitLimitEstimate returns the compute units consumed by a transaction. If there is an error, then use a hardcoded value, and send it without a bundle, send it normally.
func GetUnitLimitEstimate(client models.IClient, wallet *solana.Wallet, blockHash *solana.Hash, ix []solana.Instruction) (uint32, error) {
	computeLimitIX, err := ComputeBudgetLimit(1_400_000)
	//set the compute limit to the max
	if err != nil {
		return 0, err
	}
	//add the compute limitIX to the beginning
	ix = append([]solana.Instruction{computeLimitIX.Build()}, ix...)
	slog.Debug("Will simulate transaction soon")
	var tx *solana.Transaction
	if len(ix) > 0 {
		if blockHash == nil {
			return 0, errors.New("block hash is nil")
		}
		if ix == nil {
			return 0, errors.New("no instructions provided")
		}

		if ix[0] == nil {
			return 0, errors.New("no instructions provided")
		}
		if len(ix) > 0 && ix[1] == nil {
			return 0, errors.New("no instructions provided 2")
		}
		tx, err = solana.NewTransaction(
			ix,
			*blockHash,
			solana.TransactionPayer(wallet.PublicKey()),
		)
		if err != nil {
			return 0, err
		}
	} else {
		return 0, errors.New("no instructions provided")
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		return &wallet.PrivateKey
	})
	if err != nil {
		return 0, err
	}

	slog.Debug("Will simulate transaction now 1")
	//simulate the transaction
	resp, err := client.SimulateTransactionWithOpts(context.Background(), tx, &rpc.SimulateTransactionOpts{
		SigVerify:              true,
		ReplaceRecentBlockhash: false,
	})
	if err != nil || resp == nil {
		if err != nil {
			slog.Error("Error simulating transaction: " + err.Error())
		}
		slog.Debug(spew.Sdump(resp))
		slog.Error(tx)
		return 0, err
	}

	if resp.Value.Err != nil {
		slog.Debug("Error in simulation: " + spew.Sdump(resp.Value.Err))
		slog.Debug(spew.Sdump(resp))
		slog.Debug(tx)
		errType, errString := GetSimulationErrorType(resp.Value)
		if errType == CriticalError {
			// Critical Error detected cant continue.
			slog.Error("Critical Error detected: " + errString)
			return 0, errors.New(errString)
		}
		if errType == IgnoreableError {
			// Ignoreable error detected, continue with the transaction
			slog.Debug("Ignoreable Error detected: " + errString)
			return 0, errors.New(errString)
		}
		if errType == NoError {
			// Do nothing, continue with the transaction
		} else {
			// Unknown error detected, cant continue.
			return 0, errors.New(spew.Sprintf("Unknown error detected: %v", resp.Value.Err))
		}
	}
	slog.Debug("Got the compute units consumed")
	slog.Debug("Compute units consumed: ", *resp.Value.UnitsConsumed)
	//get the compute units consumed
	computeUnitsConsumed := resp.Value.UnitsConsumed
	return uint32(*computeUnitsConsumed), nil
}

func ConstructCreateATAInstruction(
	userPublicKey string,
	tokenMint string,
) (solana.Instruction, error) {
	if userPublicKey == "" || tokenMint == "" {
		return nil, errors.New("missing required fields")
	}

	userPublicKeyKey, err := solana.PublicKeyFromBase58(userPublicKey)
	tokenMintKey, err := solana.PublicKeyFromBase58(tokenMint)
	if err != nil {
		return nil, err

	}
	ata, _, err := common.FindAssociatedTokenAddress(common.PublicKey(userPublicKeyKey), common.PublicKey(tokenMintKey))
	//create idempotent associated token account+
	createAtaIx := associated_token_account.CreateIdempotent(associated_token_account.CreateIdempotentParam{
		Funder:                 common.PublicKey(userPublicKeyKey),
		Owner:                  common.PublicKey(userPublicKeyKey),
		Mint:                   common.PublicKey(tokenMintKey),
		AssociatedTokenAccount: ata,
	})
	var accountMetaSlice solana.AccountMetaSlice
	for _, acc := range createAtaIx.Accounts {
		accountMetaSlice.Append(&solana.AccountMeta{
			PublicKey:  solana.PublicKey(acc.PubKey),
			IsWritable: acc.IsWritable,
			IsSigner:   acc.IsSigner,
		})
	}

	programId := solana.PublicKey(createAtaIx.ProgramID)
	defaultInstruction := solana.NewInstruction(programId, accountMetaSlice, createAtaIx.Data)
	//convert to solana instruction
	return defaultInstruction, nil
}

func ComputeBudgetPrice(amount uint64) (computebudget.SetComputeUnitPrice, error) {
	ix := computebudget.NewSetComputeUnitPriceInstruction(amount)
	if ix == nil {
		return computebudget.SetComputeUnitPrice{}, errors.New("error creating compute unit price")
	}
	return *ix, nil
}

func ComputeBudgetLimit(amount uint32) (computebudget.SetComputeUnitLimit, error) {
	ix := computebudget.NewSetComputeUnitLimitInstruction(amount)
	if ix == nil {
		return computebudget.SetComputeUnitLimit{}, errors.New("error creating compute unit limit")
	}
	return *ix, nil
}

type RequestBodyPriorityFee struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      string `json:"id"`
	Method  string `json:"method"`
	Params  []struct {
		AccountKeys []string        `json:"accountKeys"`
		Options     map[string]bool `json:"options"`
	} `json:"params"`
}

type PriorityFeeLevels struct {
	Min       float64 `json:"min"`
	Low       float64 `json:"low"`
	Medium    float64 `json:"medium"`
	High      float64 `json:"high"`
	VeryHigh  float64 `json:"veryHigh"`
	UnsafeMax float64 `json:"unsafeMax"`
}

type Result struct {
	PriorityFeeLevels PriorityFeeLevels `json:"priorityFeeLevels"`
}

type ApiResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  Result `json:"result"`
	ID      string `json:"id"`
}
