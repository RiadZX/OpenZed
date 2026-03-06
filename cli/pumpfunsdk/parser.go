package pumpfunsdk

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go"
	"github.com/gookit/slog"
)

type TradeEventData struct {
	//identifier is 8 bytes
	Identifier           [8]byte
	Mint                 solana.PublicKey
	SolAmount            uint64
	TokenAmount          uint64
	IsBuy                bool
	User                 solana.PublicKey
	Timestamp            int64
	VirtualSolReserves   uint64
	VirtualTokenReserves uint64
	RealSolReserves      uint64
	RealTokenReserves    uint64
}

// ParseLog parses the logs of a transaction and returns useful information.
func ParseLog(logs []string) (handle bool, txType TransactionType, trdEventData *TradeEventData) {
	var tradeData TradeEventData
	if logs == nil || len(logs) == 0 {
		return false, Error, nil
	}

	var hasBuy = false
	var hasSell = false
	var newTokenCreated = false
	var isPumpfun = false
	var migrate = false

	for i, l := range logs {
		if len(l) >= 32 && l[:32] == "Program log: Instruction: Create" {
			if i+1 < len(logs) && len(logs[i+1]) >= 47 && logs[i+1][:47] == "Program 11111111111111111111111111111111 invoke" {
				newTokenCreated = true
			}
			continue
		}

		if len(l) >= 34 && l[:34] == "Program log: Instruction: Withdraw" {
			if i+1 < len(logs) && len(logs[i+1]) >= 58 && logs[i+1][:58] == "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke" {
				migrate = true
				break
			}
		}

		if len(l) >= 30 && l[:30] == "Program log: Instruction: Sell" {
			hasSell = true
			continue
		}

		if len(l) >= 29 && l[:29] == "Program log: Instruction: Buy" {
			hasBuy = true
			continue
		}

		if len(l) >= 59 && l[:59] == "Program 6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P success" {
			isPumpfun = true
			continue
		}

		if len(l) >= 18 && l[:18] == "Program data: vdt/" && len(logs[i+1]) >= 51 && logs[i+1][:51] == "Program 6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P" {

			data, err := decodeData(l[14:])
			if err != nil {
				slog.Error("Error decoding data: ", err)
				slog.Debug(spew.Sdump(l))
				return false, Error, nil
			}
			tradeData, err = binaryToTradeTransactionData(data)
			if err != nil {
				slog.Error("Error decoding data 2: ", err)
				slog.Debug(spew.Sdump(l))
				return false, Error, nil
			}
		}
	}
	if !isPumpfun {
		return false, NotPumpFun, nil
	}
	if migrate {
		return false, Migration, nil
	}
	if newTokenCreated && !hasBuy {
		return false, Create, nil
	}
	if newTokenCreated && hasBuy {
		return false, CreateBuy, nil
	}
	if hasBuy && hasSell {
		return false, BuySell, &tradeData
	}
	if hasSell && !hasBuy {
		return true, Sell, &tradeData
	}
	if hasBuy && !hasSell {
		return true, Buy, &tradeData
	}
	return false, Error, nil
}

func decodeData(data string) ([]byte, error) {
	if data == "" {
		return []byte{}, fmt.Errorf("data is empty")
	}
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return []byte{}, err
	}
	return decoded, nil
}
func binaryToTradeTransactionData(data []byte) (TradeEventData, error) {
	if len(data) <= 8 { //8 is just a random number.
		return TradeEventData{}, fmt.Errorf("data length is invalid")
	}
	var transactionData TradeEventData
	reader := bytes.NewReader(data)
	err := binary.Read(reader, binary.LittleEndian, &transactionData)
	if err != nil {
		return TradeEventData{}, err
	}
	return transactionData, nil
}
