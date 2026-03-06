package pumpfunsdk

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/gagliardetto/solana-go"
)

/*
	AMMBuyEvent

{"name": "timestamp"                           , "type": "i64"   },

	{"name": "base_amount_out"                     , "type": "u64"   },
	{"name": "max_quote_amount_in"                 , "type": "u64"   },
	{"name": "user_base_token_reserves"            , "type": "u64"   },
	{"name": "user_quote_token_reserves"           , "type": "u64"   },
	{"name": "pool_base_token_reserves"            , "type": "u64"   },
	{"name": "pool_quote_token_reserves"           , "type": "u64"   },
	{"name": "quote_amount_in"                     , "type": "u64"   },
	{"name": "lp_fee_basis_points"                 , "type": "u64"   },
	{"name": "lp_fee"                              , "type": "u64"   },
	{"name": "protocol_fee_basis_points"           , "type": "u64"   },
	{"name": "protocol_fee"                        , "type": "u64"   },
	{"name": "quote_amount_in_with_lp_fee"         , "type": "u64"   },
	{"name": "user_quote_amount_in"                , "type": "u64"   },
	{"name": "pool"                                , "type": "pubkey"},
	{"name": "user"                                , "type": "pubkey"},
	{"name": "user_base_token_account"             , "type": "pubkey"},
	{"name": "user_quote_token_account"            , "type": "pubkey"},
	{"name": "protocol_fee_recipient"              , "type": "pubkey"},
	{"name": "protocol_fee_recipient_token_account", "type": "pubkey"}
*/
type AMMBuyEvent struct {
	Identifier                           [8]byte
	Timestamp                            int64
	Base_amount_out                      uint64
	Max_quote_amount_in                  uint64
	User_base_token_reserves             uint64
	User_quote_token_reserves            uint64
	Pool_base_token_reserves             uint64
	Pool_quote_token_reserves            uint64
	Quote_amount_in                      uint64
	Lp_fee_basis_points                  uint64
	Lp_fee                               uint64
	Protocol_fee_basis_points            uint64
	Protocol_fee                         uint64
	Quote_amount_in_with_lp_fee          uint64
	User_quote_amount_in                 uint64
	Pool                                 solana.PublicKey
	User                                 solana.PublicKey
	User_base_token_account              solana.PublicKey
	User_quote_token_account             solana.PublicKey
	Protocol_fee_recipient               solana.PublicKey
	Protocol_fee_recipient_token_account solana.PublicKey
}

/*
AMMSellEvent

{"name": "timestamp"                           , "type": "i64"   },
{"name": "base_amount_in"                      , "type": "u64"   },
{"name": "min_quote_amount_out"                , "type": "u64"   },
{"name": "user_base_token_reserves"            , "type": "u64"   },
{"name": "user_quote_token_reserves"           , "type": "u64"   },
{"name": "pool_base_token_reserves"            , "type": "u64"   },
{"name": "pool_quote_token_reserves"           , "type": "u64"   },
{"name": "quote_amount_out"                    , "type": "u64"   },
{"name": "lp_fee_basis_points"                 , "type": "u64"   },
{"name": "lp_fee"                              , "type": "u64"   },
{"name": "protocol_fee_basis_points"           , "type": "u64"   },
{"name": "protocol_fee"                        , "type": "u64"   },
{"name": "quote_amount_out_without_lp_fee"     , "type": "u64"   },
{"name": "user_quote_amount_out"               , "type": "u64"   },
{"name": "pool"                                , "type": "pubkey"},
{"name": "user"                                , "type": "pubkey"},
{"name": "user_base_token_account"             , "type": "pubkey"},
{"name": "user_quote_token_account"            , "type": "pubkey"},
{"name": "protocol_fee_recipient"              , "type": "pubkey"},
{"name": "protocol_fee_recipient_token_account", "type": "pubkey"}
*/
type AMMSellEvent struct {
	Identifier                           [8]byte
	Timestamp                            int64
	Base_amount_in                       uint64
	Min_quote_amount_out                 uint64
	User_base_token_reserves             uint64
	User_quote_token_reserves            uint64
	Pool_base_token_reserves             uint64
	Pool_quote_token_reserves            uint64
	Quote_amount_out                     uint64
	Lp_fee_basis_points                  uint64
	Lp_fee                               uint64
	Protocol_fee_basis_points            uint64
	Protocol_fee                         uint64
	Quote_amount_out_without_lp_fee      uint64
	User_quote_amount_out                uint64
	Pool                                 solana.PublicKey
	User                                 solana.PublicKey
	User_base_token_account              solana.PublicKey
	User_quote_token_account             solana.PublicKey
	Protocol_fee_recipient               solana.PublicKey
	Protocol_fee_recipient_token_account solana.PublicKey
}

// AMMEvent is a struct that contains either a Buy or a Sell event.
// It is used to parse the logs of a transaction.
type AMMEvent struct {
	IsBuy bool
	Buy   *AMMBuyEvent
	Sell  *AMMSellEvent
}

// ParseAMMLog parses the logs of a transaction and returns useful information.
func ParseAMMLog(logs []string) (txType TransactionType, _ammEvent *AMMEvent) {
	var ammEvent AMMEvent
	if logs == nil || len(logs) == 0 {
		return Error, nil
	}

	for i, l := range logs {
		if len(l) >= 60 && l[:60] == "Program pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA invoke [" {
			if i+1 < len(logs) && len(logs[i+1]) >= 30 && logs[i+1][:30] == "Program log: Instruction: Sell" {
				ammEvent.IsBuy = false
				ammEvent.Sell = &AMMSellEvent{}
				continue
			}
		}

		if len(l) >= 60 && l[:60] == "Program pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA invoke [" {
			if i+1 < len(logs) && len(logs[i+1]) >= 29 && logs[i+1][:29] == "Program log: Instruction: Buy" {
				ammEvent.IsBuy = true
				ammEvent.Buy = &AMMBuyEvent{}
				continue
			}
		}

		// finding the program data
		if len(l) >= 59 && l[:59] == "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success" {
			if i+1 < len(logs) && len(logs[i+1]) >= 13 && logs[i+1][:13] == "Program data:" {
				if i+2 < len(logs) && len(logs[i+2]) >= 60 && logs[i+2][:60] == "Program pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA invoke [" {
					// Get the program data
					data, err := decodeData(logs[i+1][14:])
					if err != nil {
						return Error, nil
					}

					if ammEvent.IsBuy {
						// Parse the data
						ammEvent.Buy, err = binaryToAMMBuyEvent(data)
						if err != nil {
							return Error, nil
						}

						return AMMBuy, &ammEvent
					} else {
						// Parse the data
						ammEvent.Sell, err = binaryToAMMSellEvent(data)
						if err != nil {
							return Error, nil
						}

						return AMMSell, &ammEvent
					}

				}
			}
		}

	}

	return Error, nil
}

func binaryToAMMBuyEvent(data []byte) (*AMMBuyEvent, error) {
	if len(data) < 8+8+8+8+8+8+8+8+8+8+8+8+8+8+8+32+32+32+32+32+32 {
		return nil, fmt.Errorf("not enough data to decode")
	}

	var ammBuyEvent AMMBuyEvent
	// Decode the data
	reader := bytes.NewReader(data)
	err := binary.Read(reader, binary.LittleEndian, &ammBuyEvent)
	if err != nil {
		return nil, err
	}

	return &ammBuyEvent, nil
}

func binaryToAMMSellEvent(data []byte) (*AMMSellEvent, error) {
	if len(data) < 8+8+8+8+8+8+8+8+8+8+8+8+8+8+8+32+32+32+32+32+32 {
		return nil, fmt.Errorf("not enough data to decode")
	}

	var ammSellEvent AMMSellEvent
	// Decode the data
	reader := bytes.NewReader(data)
	err := binary.Read(reader, binary.LittleEndian, &ammSellEvent)
	if err != nil {
		return nil, err
	}

	return &ammSellEvent, nil
}
