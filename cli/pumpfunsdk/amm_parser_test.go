package pumpfunsdk

import (
	"github.com/gagliardetto/solana-go"
	"testing"
)

var (
	expectedBuyEvent = &AMMBuyEvent{
		Identifier:                           [8]byte{103, 244, 82, 31, 44, 245, 119, 119},
		Timestamp:                            1742548782,
		Base_amount_out:                      1004602411388,
		Max_quote_amount_in:                  495000001,
		User_base_token_reserves:             0,
		User_quote_token_reserves:            500000001,
		Pool_base_token_reserves:             189886878455672,
		Pool_quote_token_reserves:            92836296886,
		Quote_amount_in:                      493765586,
		Lp_fee_basis_points:                  20,
		Lp_fee:                               987532,
		Protocol_fee_basis_points:            5,
		Protocol_fee:                         246883,
		Quote_amount_in_with_lp_fee:          494753118,
		User_quote_amount_in:                 495000001,
		Pool:                                 solana.MustPublicKeyFromBase58("8oKCjDecqf1yGYgwpVAZejnDox7c6kswE82Xx1yHBhnk"),
		User:                                 solana.MustPublicKeyFromBase58("4QTLg5ChZxs3kDGK2JAxbkxPgxgnBAKA5JMuL87J9L2P"),
		User_base_token_account:              solana.MustPublicKeyFromBase58("26f3QQ58N9hhHxB9YEj3sPjDDBD6kxr2fGVXKFXbfJ2S"),
		User_quote_token_account:             solana.MustPublicKeyFromBase58("CULQFN9tZob617pXFztYCCzqzyGS1jPgd2BocbZskw8j"),
		Protocol_fee_recipient:               solana.MustPublicKeyFromBase58("JCRGumoE9Qi5BBgULTgdgTLjSgkCMSbF62ZZfGs84JeU"),
		Protocol_fee_recipient_token_account: solana.MustPublicKeyFromBase58("DWpvfqzGWuVy9jVSKSShdM2733nrEsnnhsUStYbkj6Nn"),
	}

	expectedSellEvent = &AMMSellEvent{
		Identifier:                           [8]byte{62, 47, 55, 10, 165, 3, 220, 42},
		Timestamp:                            1742500208,
		Base_amount_in:                       24280378766,
		Min_quote_amount_out:                 7151622,
		User_base_token_reserves:             24280378766,
		User_quote_token_reserves:            0,
		Pool_base_token_reserves:             222398710005666,
		Pool_quote_token_reserves:            79114125035,
		Quote_amount_out:                     8636341,
		Lp_fee_basis_points:                  20,
		Lp_fee:                               17273,
		Protocol_fee_basis_points:            5,
		Protocol_fee:                         4319,
		Quote_amount_out_without_lp_fee:      8619068,
		User_quote_amount_out:                8614749,
		Pool:                                 solana.MustPublicKeyFromBase58("FZfk5akAiF1StjSLshXRVYyKUKTmEKsU3fQRHnProEGr"),
		User:                                 solana.MustPublicKeyFromBase58("9RyAKViy6kGmMoi4qXttFdugMYzDYJdvHfEpcoyKXBeM"),
		User_base_token_account:              solana.MustPublicKeyFromBase58("B5oFU8nACVPMNFiutq3m7VAQh3NcYXZo9kDWhNBuAvMq"),
		User_quote_token_account:             solana.MustPublicKeyFromBase58("CdH8MTfM3uXxWyonwYz65DSja5R73W4aNp7hMc2p8u45"),
		Protocol_fee_recipient:               solana.MustPublicKeyFromBase58("7hTckgnGnLQR6sdH7YkqFTAA7VwTfYFaZ6EhEsU3saCX"),
		Protocol_fee_recipient_token_account: solana.MustPublicKeyFromBase58("X5QPJcpph4mBAJDzc4hRziFftSbcygV59kRb2Fu6Je1"),
	}
)

func TestParseAMMLog(t *testing.T) {
	// Test sell log
	shouldProcess, transactionType, ammEvent := ParseAMMLog(sellLogAMM)
	if !shouldProcess {
		t.Errorf("Expected to process log, got: %v", shouldProcess)
	}
	if transactionType != AMMSell {
		t.Errorf("Expected transaction type to be Sell, got: %v", transactionType)
	}
	if ammEvent == nil || ammEvent.IsBuy {
		t.Errorf("Expected trade data to be non-nil, got: %v", ammEvent)
	}
	if *ammEvent.Sell != *expectedSellEvent {
		t.Errorf("Expected trade data: %+v, got: %+v", expectedSellEvent, ammEvent.Sell)
	}

	// Test buy log
	shouldProcess, transactionType, ammEvent = ParseAMMLog(buyLogAMM)
	if !shouldProcess {
		t.Errorf("Expected to process log, got: %v", shouldProcess)
	}
	if transactionType != AMMBuy {
		t.Errorf("Expected transaction type to be Buy, got: %v", transactionType)
	}
	if ammEvent == nil || !ammEvent.IsBuy {
		t.Errorf("Expected trade data to be non-nil, got: %v", ammEvent)
	}
	if *ammEvent.Buy != *expectedBuyEvent {
		t.Errorf("Expected trade data: %+v, got: %+v", expectedBuyEvent, ammEvent.Buy)
	}

}

func TestParseAMMLog2(t *testing.T) {
	// THIS SHOULD NOT PARSE NORMAL PF LOGS. SO IT SHOULD FAIL ON ALL.
	data := "Pi83CqUD3CpwcdxnAAAAAI4tOacFAAAABiBtAAAAAACOLTmnBQAAAAAAAAAAAAAAor8bPEXKAADrvpFrEgAAALXHgwAAAAAAFAAAAAAAAAB5QwAAAAAAAAUAAAAAAAAA3xAAAAAAAAA8hIMAAAAAAF1zgwAAAAAA2GICaSBcjvYyhpap2jRjJNRArZo5NUf3pdLOtVAUZKt9Qu+XNrjSBor8vae0hg7n648Hi26/eL/NBmjdha0arpXPfKZrymyxPap9pdiD5Npyoaq1sNkD39drGouX8zBUrLuzLac41f0NeWasYTn1ibV2tnD1El0ZqdT5nKSKo2pjg3MADqIssmTTSv9koEte+r+7dN3NBImXsZgVR9fREAe0ZyjFA6fIFZjsUWe5tjKg2nvc6Y8HxZZ7EO1veKHO"
	tests := []struct {
		name     string
		logs     []string
		expected TransactionType
	}{
		{"create", []string{"Program log: Instruction: Create", "Program 11111111111111111111111111111111 invoke", "Program log: Instruction: Sell", "Program 6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P success"}, Create},
		{"sell", []string{"Program log: Instruction: Sell", data, "Program 6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P success"}, Sell},
		{"buy", []string{"Program log: Instruction: Buy", data, "Program 6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P success"}, Buy},
		{"buy_sell", []string{"Program log: Instruction: Buy", data, "Program 6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P success", "Program log: Instruction: Sell", "Program 6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P success"}, BuySell},
		{"create_buy", []string{"Program log: Instruction: Create", "Program 11111111111111111111111111111111 invoke", "Program log: Instruction: Buy", "Program 6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P success"}, CreateBuy},
		{"error", []string{"Program skibi"}, NotPumpFun},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handle, _, tradeData := ParseAMMLog(tt.logs)
			if handle || tradeData != nil {
				t.Errorf("expected handle to be false")
			}
		})
	}

}

func TestDecoderAMMSellEvent(t *testing.T) {
	//https://solscan.io/tx/3W6ybgQPz3qq1XFijcH9GzBThb2oyS4NSDmyfwoioKngBPwnaFg3C3waShwuee6QtfjLcGZpForZy9jxHHUsD7BE
	data := "Pi83CqUD3CpwcdxnAAAAAI4tOacFAAAABiBtAAAAAACOLTmnBQAAAAAAAAAAAAAAor8bPEXKAADrvpFrEgAAALXHgwAAAAAAFAAAAAAAAAB5QwAAAAAAAAUAAAAAAAAA3xAAAAAAAAA8hIMAAAAAAF1zgwAAAAAA2GICaSBcjvYyhpap2jRjJNRArZo5NUf3pdLOtVAUZKt9Qu+XNrjSBor8vae0hg7n648Hi26/eL/NBmjdha0arpXPfKZrymyxPap9pdiD5Npyoaq1sNkD39drGouX8zBUrLuzLac41f0NeWasYTn1ibV2tnD1El0ZqdT5nKSKo2pjg3MADqIssmTTSv9koEte+r+7dN3NBImXsZgVR9fREAe0ZyjFA6fIFZjsUWe5tjKg2nvc6Y8HxZZ7EO1veKHO"
	decoded, err := decodeData(data)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	transactionData, err := binaryToAMMSellEvent(decoded)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if *transactionData != *expectedSellEvent {
		t.Errorf("Expected transaction data: %+v, got: %+v", expectedSellEvent, transactionData)
	}

}

func TestDecoderAMMBuyEvent(t *testing.T) {
	//https://solscan.io/tx/4b7u1PevsBroBvfxtqJCEXQv2RAzgrFPEaj2oibbmCCLFVLFFDrpJmibCZmMBf82tDo9idDkoKmY2gVNaAqu48t4
	data := "Z/RSHyz1d3cuL91nAAAAAHxJ+ObpAAAAwRmBHQAAAAAAAAAAAAAAAAFlzR0AAAAAeBfwe7OsAAC2eHmdFQAAANJDbh0AAAAAFAAAAAAAAACMEQ8AAAAAAAUAAAAAAAAAY8QDAAAAAABeVX0dAAAAAMEZgR0AAAAAc98bZGjiqtvkfAhFNtcVsZKhZjBhc/jUJLNFh+rrxw0ylTOC+61M3dPOnnlgOZmS032iW1Iu1FSBx4FDwI0OTBBOmsSissG41NmeFbKwTzcUFcxTJrcJg/XWzl52jQ0HqnEss4KRS++BiSiUZcUVdAPsjOBkMxTM1YedXeIfVzD/g4OBi6j6KMPNO21ek/n6uPCXm8NyFazFskaHe6jDybnwRqOiz07iErT6kv2dAPpPsLP8/TYeQK5TkflgNepj"
	decoded, err := decodeData(data)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	transactionData, err := binaryToAMMBuyEvent(decoded)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if *transactionData != *expectedBuyEvent {
		t.Errorf("Expected transaction data: %+v, got: %+v", expectedBuyEvent, transactionData)
	}

}

var sellLogAMM = []string{
	"Program ComputeBudget111111111111111111111111111111 invoke [1]",
	"Program ComputeBudget111111111111111111111111111111 success",
	"Program ComputeBudget111111111111111111111111111111 invoke [1]",
	"Program ComputeBudget111111111111111111111111111111 success",
	"Program 11111111111111111111111111111111 invoke [1]",
	"Program 11111111111111111111111111111111 success",
	"Program ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL invoke [1]",
	"Program log: CreateIdempotent",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
	"Program log: Instruction: GetAccountDataSize",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 1569 of 79817 compute units",
	"Program return: TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA pQAAAAAAAAA=",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
	"Program 11111111111111111111111111111111 invoke [2]",
	"Program 11111111111111111111111111111111 success",
	"Program log: Initialize the associated token account",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
	"Program log: Instruction: InitializeImmutableOwner",
	"Program log: Please upgrade to SPL Token 2022 for immutable owner support",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 1405 of 73230 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
	"Program log: Instruction: InitializeAccount3",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 3158 of 69348 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
	"Program ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL consumed 19315 of 85222 compute units",
	"Program ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL success",
	"Program pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA invoke [1]",
	"Program log: Instruction: Sell",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
	"Program log: Instruction: TransferChecked",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 6147 of 39549 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
	"Program log: Instruction: TransferChecked",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 6238 of 30654 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
	"Program log: Instruction: TransferChecked",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 6238 of 21655 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
	"Program data: Pi83CqUD3CpwcdxnAAAAAI4tOacFAAAABiBtAAAAAACOLTmnBQAAAAAAAAAAAAAAor8bPEXKAADrvpFrEgAAALXHgwAAAAAAFAAAAAAAAAB5QwAAAAAAAAUAAAAAAAAA3xAAAAAAAAA8hIMAAAAAAF1zgwAAAAAA2GICaSBcjvYyhpap2jRjJNRArZo5NUf3pdLOtVAUZKt9Qu+XNrjSBor8vae0hg7n648Hi26/eL/NBmjdha0arpXPfKZrymyxPap9pdiD5Npyoaq1sNkD39drGouX8zBUrLuzLac41f0NeWasYTn1ibV2tnD1El0ZqdT5nKSKo2pjg3MADqIssmTTSv9koEte+r+7dN3NBImXsZgVR9fREAe0ZyjFA6fIFZjsUWe5tjKg2nvc6Y8HxZZ7EO1veKHO",
	"Program pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA invoke [2]",
	"Program pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA consumed 2004 of 9538 compute units",
	"Program pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA success",
	"Program pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA consumed 58912 of 65907 compute units",
	"Program pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA success",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [1]",
	"Program log: Instruction: CloseAccount",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 2915 of 6995 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
}

var buyLogAMM = []string{
	"Program ComputeBudget111111111111111111111111111111 invoke [1]",
	"Program ComputeBudget111111111111111111111111111111 success",
	"Program ComputeBudget111111111111111111111111111111 invoke [1]",
	"Program ComputeBudget111111111111111111111111111111 success",
	"Program 11111111111111111111111111111111 invoke [1]",
	"Program 11111111111111111111111111111111 success",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [1]",
	"Program log: Instruction: InitializeAccount",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 3443 of 119550 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
	"Program ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL invoke [1]",
	"Program log: CreateIdempotent",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
	"Program log: Instruction: GetAccountDataSize",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 1569 of 110702 compute units",
	"Program return: TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA pQAAAAAAAAA=",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
	"Program 11111111111111111111111111111111 invoke [2]",
	"Program 11111111111111111111111111111111 success",
	"Program log: Initialize the associated token account",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
	"Program log: Instruction: InitializeImmutableOwner",
	"Program log: Please upgrade to SPL Token 2022 for immutable owner support",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 1405 of 104115 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
	"Program log: Instruction: InitializeAccount3",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4188 of 100233 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
	"Program ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL consumed 20345 of 116107 compute units",
	"Program ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL success",
	"Program HgoHJy31rnpmm99CaoKn72g1QDLf6A8vzqEKAXCyBFv5 invoke [1]",
	"Program pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA invoke [2]",
	"Program log: Instruction: Buy",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [3]",
	"Program log: Instruction: TransferChecked",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 6147 of 64494 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [3]",
	"Program log: Instruction: TransferChecked",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 6238 of 55680 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [3]",
	"Program log: Instruction: TransferChecked",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 6238 of 46790 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
	"Program data: Z/RSHyz1d3cuL91nAAAAAHxJ+ObpAAAAwRmBHQAAAAAAAAAAAAAAAAFlzR0AAAAAeBfwe7OsAAC2eHmdFQAAANJDbh0AAAAAFAAAAAAAAACMEQ8AAAAAAAUAAAAAAAAAY8QDAAAAAABeVX0dAAAAAMEZgR0AAAAAc98bZGjiqtvkfAhFNtcVsZKhZjBhc/jUJLNFh+rrxw0ylTOC+61M3dPOnnlgOZmS032iW1Iu1FSBx4FDwI0OTBBOmsSissG41NmeFbKwTzcUFcxTJrcJg/XWzl52jQ0HqnEss4KRS++BiSiUZcUVdAPsjOBkMxTM1YedXeIfVzD/g4OBi6j6KMPNO21ek/n6uPCXm8NyFazFskaHe6jDybnwRqOiz07iErT6kv2dAPpPsLP8/TYeQK5TkflgNepj",
	"Program pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA invoke [3]",
	"Program pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA consumed 2004 of 34688 compute units",
	"Program pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA success",
	"Program pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA consumed 58889 of 91034 compute units",
	"Program pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA success",
	"Program HgoHJy31rnpmm99CaoKn72g1QDLf6A8vzqEKAXCyBFv5 consumed 63684 of 95762 compute units",
	"Program HgoHJy31rnpmm99CaoKn72g1QDLf6A8vzqEKAXCyBFv5 success",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [1]",
	"Program log: Instruction: CloseAccount",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 2915 of 32078 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
	"Program 11111111111111111111111111111111 invoke [1]",
	"Program 11111111111111111111111111111111 success",
}
