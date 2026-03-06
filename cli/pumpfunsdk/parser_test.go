package pumpfunsdk

import (
	"github.com/gagliardetto/solana-go"
	"reflect"
	"testing"
)

func TestParseLog(t *testing.T) {
	data := "Program data: vdt/007mYe6O/fZjqHQnDCflfHjnMaMO3MOlS+QQHe+irZicPP1Db+DDxR0AAAAAobzMfxEJAAAB44R4ltbau0ov3A+xdT+p4huWz3Odj0sSUrP9MTbZIFlc6H9mAAAAAF7WfWgJAAAA5kQ8k4fUAgBeKlpsAgAAAOasKUf21QEA"
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
			_, txType, tradeData := ParseLog(tt.logs)
			//fail and print expected vs actual
			if txType != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, txType)
			}

			if txType == Buy || txType == Sell || txType == BuySell {
				if tradeData == nil || tradeData.SolAmount == 0 || tradeData.TokenAmount == 0 {
					t.Errorf("expected tradeData to not be nil")
				}
			}

			if txType == Create || txType == CreateBuy {
				if tradeData != nil {
					t.Errorf("expected tradeData to be nil")
				}
			}
		})
	}

}

func TestDecodeTradeEventData(t *testing.T) {
	//https://solana.fm/tx/48rPayzmA3E67ju6fp2QvTPcqC9hXhs5wFavNjMAbNfvpsb1Pzgsw9eVv1T5nAavnncQ4whrJYayyaHyUwSGgZgC
	data := "vdt/007mYe7z81ZQ4IRFJrN2mfy5cR33GIKmiWNV1aW9P4TwbdXR/wBe0LIAAAAAX+JrA0wKAAABsj79Pa7eAciGQmjG6zqRHKoMer1r4BweBTLvXRx0DitbKH9mAAAAAJEdAtsVAAAAn3ZXq+M3AQCRcd7eDgAAAJ/eRF9SOQAA"
	decoded, err := decodeData(data)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	transactionData, err := binaryToTradeTransactionData(decoded)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	expectedIdentifier := [8]byte{189, 219, 127, 211, 78, 230, 97, 238}
	if transactionData.Identifier != expectedIdentifier {
		t.Errorf("Expected identifier: %v, got: %v", expectedIdentifier, transactionData.Identifier)
	}
	expectedMint := solana.MustPublicKeyFromBase58("HRHHCBAmrHJf95K2uk9c6dYPt5RNHcToDRvpGTo8pump")
	if transactionData.Mint != expectedMint {
		t.Errorf("Expected mint: %v, got: %v", expectedMint, transactionData.Mint)
	}
	expectedSolAmount := uint64(3000000000)
	if transactionData.SolAmount != expectedSolAmount {
		t.Errorf("Expected sol amount: %v, got: %v", expectedSolAmount, transactionData.SolAmount)
	}
	expectedTokenAmount := uint64(11321591194207)
	if transactionData.TokenAmount != expectedTokenAmount {
		t.Errorf("Expected token amount: %v, got: %v", expectedTokenAmount, transactionData.TokenAmount)
	}

	if transactionData.IsBuy != true {
		t.Errorf("Expected is buy: %v, got: %v", true, transactionData.IsBuy)
	}
	expectedUser := solana.MustPublicKeyFromBase58("CzoH8q3Xmg96EUrxtFxh68n7srHE4wcG7WbJvwLnFTxn")
	if transactionData.User != expectedUser {
		t.Errorf("Expected user: %v, got: %v", expectedUser, transactionData.User)
	}
	expectedTimestamp := int64(1719609435)
	if transactionData.Timestamp != expectedTimestamp {
		t.Errorf("Expected timestamp: %v, got: %v", expectedTimestamp, transactionData.Timestamp)
	}
	expectedVirtualSolReserves := uint64(93868662161)
	if transactionData.VirtualSolReserves != expectedVirtualSolReserves {
		t.Errorf("Expected virtual sol reserves: %v, got: %v", expectedVirtualSolReserves, transactionData.VirtualSolReserves)
	}
	expectedVirtualTokenReserves := uint64(342925948450463)
	if transactionData.VirtualTokenReserves != expectedVirtualTokenReserves {
		t.Errorf("Expected virtual token reserves: %v, got: %v", expectedVirtualTokenReserves, transactionData.VirtualTokenReserves)
	}
	expectedRealSolReserves := uint64(63868662161)
	if transactionData.RealSolReserves != expectedRealSolReserves {
		t.Errorf("Expected real sol reserves: %v, got: %v", expectedRealSolReserves, transactionData.RealSolReserves)
	}
	expectedRealTokenReserves := uint64(63025948450463)
	if transactionData.RealTokenReserves != expectedRealTokenReserves {
		t.Errorf("Expected real token reserves: %v, got: %v", expectedRealTokenReserves, transactionData.RealTokenReserves)
	}
}

func TestDecodeEvent(t *testing.T) {
	data := "vdt/007mYe6O/fZjqHQnDCflfHjnMaMO3MOlS+QQHe+irZicPP1Db+DDxR0AAAAAobzMfxEJAAAB44R4ltbau0ov3A+xdT+p4huWz3Odj0sSUrP9MTbZIFlc6H9mAAAAAF7WfWgJAAAA5kQ8k4fUAgBeKlpsAgAAAOasKUf21QEA"
	decoded, err := decodeData(data)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	transactionData, err := binaryToTradeTransactionData(decoded)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	expectedIdentifier := [8]byte{189, 219, 127, 211, 78, 230, 97, 238}
	if transactionData.Identifier != expectedIdentifier {
		t.Errorf("Expected identifier: %v, got: %v", expectedIdentifier, transactionData.Identifier)
	}
	//print the decoded data
	if transactionData.SolAmount != 499500000 {
		t.Fatal("Expected SolAmount: 499500000, got: ", transactionData.SolAmount)
	}
}

func TestDecodeEvent2(t *testing.T) {
	data := "vdt/007mYe6O/fZjqHQnDCflfHfwafawaMO3MOlS+QQHe+irZicPP1Db+DDxR0AAAAAobzMfxEJAAAB44R4ltbau0ov3A+xdT+p4huWz3Odj0sSUrP9MTbZIFlc6H9mAAAAAF7WfWgJAAAA5kQ8k4fUAgBeKlpsAgAAAOasKUf21QEA"
	_, err := decodeData(data)
	if err == nil {
		t.Fail()
	}
}

func TestDecodeEvent3(t *testing.T) {
	data := ""
	_, err := decodeData(data)
	if err == nil {
		t.Fail()
	}
}

func TestDecodeEvent4(t *testing.T) {
	//THIS IS NOT A PUMPFUN DATA, THIS IS INVALID DATA!!
	data := "vdt/007mYe4WmJ2ufKhShGagUrho4px+18834clcfeJV7cH7ky3l7spH6nVBk2lCczUwvAlALF4nFRPdnuqOqEwuDxQBA1QtagFpAAAAAABjF2kAAAAAAN5gtQEAAAAAZLy1AQAAAAA="
	dataX, err := decodeData(data)
	if err != nil {
		t.Fatal(err)
	}
	_, err = binaryToTradeTransactionData(dataX)
	if err == nil {
		t.Fatal(err)
	}

}

func TestBinaryToTradeTransactionData(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    TradeEventData
		wantErr bool
	}{
		{
			name: "nil data",
			args: args{
				data: nil,
			},
			want:    TradeEventData{},
			wantErr: true,
		},
		{
			name: "empty data",
			args: args{
				data: []byte{},
			},
			want:    TradeEventData{},
			wantErr: true,
		},
		{
			name: "invalid data",
			args: args{
				data: []byte("invalid data"),
			},
			want:    TradeEventData{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := binaryToTradeTransactionData(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("BinaryToTradeTransactionData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BinaryToTradeTransactionData() got = %v, want %v", got, tt.want)
			}
		})
	}
}
