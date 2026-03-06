package raydiumsdk

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/mr-tron/base58"
)

type TransactionType int

const (
	Swap TransactionType = iota
	Error
	NotRaydium
	MultipleSwaps
)

type EventData struct {
	LogType          uint8
	AmountIn         uint64
	MinimumAmountOut uint64

	Direction  uint64
	UserSource uint64
	PoolCoin   uint64
	PoolPC     uint64
	AmountOut  uint64
}

type SwapData struct {
	Discriminator    uint8
	AmountIn         uint64
	MinimumAmountOut uint64
}

// ParseLog parses the logs of a transaction and returns useful information.
func ParseLog(logs []string) (handle bool, txType TransactionType, eventData EventData) {
	if logs == nil || len(logs) == 0 {
		return false, Error, EventData{}
	}

	var evData EventData

	isRaydium := false
	isSwap := false
	isMultipleSwaps := false
	success := false
	for i, l := range logs {

		if len(l) >= 59 && l[:59] == "Program PhoeNiXZ8ByJGLkxNfZRnkUfjvmuYqLR89jjFHGqdXY success" {
			isRaydium = false
			break
		}

		if len(l) >= 51 && l[:51] == "Program JUP6LkbZbjS1jKKwapdHNy74zcZ3tLUZoi5QNyVTaV4" {
			isRaydium = false
			break
		}

		if len(l) >= 51 && l[:51] == "Program LBUZKhRxPF3XUpBCjp4YzTKgLccjZhTSDM9YuVaPwxo" {
			isRaydium = false
			break
		}

		if len(l) >= 19 && l[:19] == "Program log: @@@@@@" {
			isRaydium = false
			break
		}

		if len(l) >= 38 && l[:38] == "Program log: Instruction: CreateApeAta" {
			isRaydium = false
			break
		}

		if len(l) >= 60 && l[:60] == "Program 675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8 success" {
			success = true
		}

		if len(l) >= 21 && l[:21] == "Program log: ray_log:" {
			if i+1 < len(logs) && len(logs[i+1]) >= 51 && logs[i+1][:51] == "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA" {
				if isSwap {
					isMultipleSwaps = true
					isSwap = false
					break
				} else {
					isSwap = true
				}
				//Parse event data
				if len(l) >= 22 {
					//if last character is a , remove it
					if l[len(l)-1] == ',' {
						l = l[:len(l)-1]
					}
					data := l[22:]
					decoded, err := DecodeData(data)
					if err != nil {
						return false, Error, EventData{}
					}
					evData, err = binaryToSwapEvent(decoded)
					if err != nil {
						return false, Error, EventData{}
					}
				}
			}
		}

		if len(l) >= 60 && l[:60] == "Program 675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8 success" {
			isRaydium = true
		}
	}
	if !success {
		return false, Error, EventData{}
	}
	if !isRaydium {
		return false, NotRaydium, EventData{}
	}
	if isMultipleSwaps {
		return false, MultipleSwaps, EventData{}
	}
	if isSwap {
		return true, Swap, evData
	}
	return false, Error, EventData{}
}

var (
	Usdc = solana.MustPublicKeyFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")
	Wsol = solana.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")
)

func DecodeData(data string) ([]byte, error) {
	if data == "" {
		return []byte{}, fmt.Errorf("data is empty")
	}
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		//try decoding in base58
		decoded, err = base58.Decode(data)
		if err != nil {
			return []byte{}, err
		}
		return decoded, nil
	}
	return decoded, nil
}
func binaryToSwapEvent(data []byte) (EventData, error) {
	if len(data) <= 8 { //8 is just a random number.
		return EventData{}, fmt.Errorf("data length is invalid")
	}
	var evData EventData
	reader := bytes.NewReader(data)
	err := binary.Read(reader, binary.LittleEndian, &evData)
	if err != nil {
		return EventData{}, err
	}
	return evData, nil
}

func binaryToSwapData(data []byte) (SwapData, error) {
	if len(data) <= 8 {
		return SwapData{}, fmt.Errorf("data length is invalid")
	}
	var sData SwapData
	reader := bytes.NewReader(data)
	err := binary.Read(reader, binary.LittleEndian, &sData)
	if err != nil {
		return SwapData{}, err
	}
	return sData, nil
}
