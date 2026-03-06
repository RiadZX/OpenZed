package raydiumsdk

import (
	"bytes"
	"reflect"
	"testing"
)

func Test_ParseLogsSell(t *testing.T) {
	gotParseLogs, txType, eventData := ParseLog(logsSell)
	if gotParseLogs != true {
		t.Errorf("ParseLogs() error = %v", gotParseLogs)
		return
	}
	if txType != Swap {
		t.Errorf("ParseLogs() error = %v", txType)
	}
	if eventData.Direction != 2 {
		t.Errorf("ParseLogs() error = %v", eventData.Direction)
	}
}

func Test_ParseLogsBuy(t *testing.T) {
	gotParseLogs, txType, eventData := ParseLog(logsBuy)
	if gotParseLogs != true {
		t.Errorf("ParseLogs() error = %v", gotParseLogs)
		return
	}
	if txType != Swap {
		t.Errorf("ParseLogs() error = %v", txType)
	}
	if eventData.Direction != 1 {
		t.Errorf("ParseLogs() error = %v", eventData.Direction)
	}
}

func Test_ParseLogsInvalid(t *testing.T) {
	gotParseLogs, _, _ := ParseLog(logsInvalid)
	if gotParseLogs != false {
		t.Errorf("ParseLogs() error = %v", gotParseLogs)
		return
	}

	gotParseLogs, _, _ = ParseLog([]string{})
	if gotParseLogs != false {
		t.Errorf("ParseLogs() error = %v", gotParseLogs)
		return
	}

	gotParseLogs, _, _ = ParseLog([]string{"", "", "", "", "", ""})
	if gotParseLogs != false {
		t.Errorf("ParseLogs() error = %v", gotParseLogs)
		return
	}

	gotParseLogs, _, _ = ParseLog(nil)
	if gotParseLogs != false {
		t.Errorf("ParseLogs() error = %v", gotParseLogs)
		return
	}
}

var logsInvalid = []string{
	"Program ComputeBudget111111111111111111111111111111 invoke [1]",
	"Program ComputeBudget111111111111111111111111111111 success",
	"Program 11111111111111111111111111111111 invoke [1]",
	"Program 11111111111111111111111111111111 success",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [1]",
	"Program log: Instruction: InitializeAccount",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 3443 of 799700 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
	"Program 675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8 invoke [1]",
	"Program log: ray_log: AwAQXl",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
	"Program log: Instruction: Transfer",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4736 of 778100 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
	"Program log: Instruction: Transfer",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4645 of 770383 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
	"Program 675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8 consumed 31322 of 796257 compute units",
	"Program 675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8 success",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [1]",
	"Program log: Instruction: CloseAccount",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 2915 of 764935 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
}

var logsBuy = []string{
	"Program ComputeBudget111111111111111111111111111111 invoke [1]",
	"Program ComputeBudget111111111111111111111111111111 success",
	"Program 11111111111111111111111111111111 invoke [1]",
	"Program 11111111111111111111111111111111 success",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [1]",
	"Program log: Instruction: InitializeAccount",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 3443 of 799700 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
	"Program 675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8 invoke [1]",
	"Program log: ray_log: AwAQXl8AAAAA0WPMcQ4AAAABAAAAAAAAAAAQXl8AAAAAUgMYKNYRAABb6qfEOAAAAPUG9LEdAAAA",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
	"Program log: Instruction: Transfer",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4736 of 778100 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
	"Program log: Instruction: Transfer",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4645 of 770383 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
	"Program 675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8 consumed 31322 of 796257 compute units",
	"Program 675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8 success",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [1]",
	"Program log: Instruction: CloseAccount",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 2915 of 764935 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
}

var logsSell = []string{
	"Program ComputeBudget111111111111111111111111111111 invoke [1]",
	"Program ComputeBudget111111111111111111111111111111 success",
	"Program 11111111111111111111111111111111 invoke [1]",
	"Program 11111111111111111111111111111111 success",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [1]",
	"Program log: Instruction: InitializeAccount",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 3443 of 799700 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
	"Program 675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8 invoke [1]",
	"Program log: ray_log: A+ClmPsNAAAAW42cEAAAAAACAAAAAAAAAJwMnaIWAAAA3qQ4uUwUAACNJCqdLQAAANvaQR8AAAAA",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
	"Program log: Instruction: Transfer",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4645 of 778238 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
	"Program log: Instruction: Transfer",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4736 of 770612 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
	"Program 675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8 consumed 31183 of 796257 compute units",
	"Program 675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8 success",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [1]",
	"Program log: Instruction: CloseAccount",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 2915 of 765074 compute units",
	"Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA succes",
}

func TestDecodeData(t *testing.T) {
	data := "A8BjlzoAAAAAAAAAAAAAAAABAAAAAAAAAMBjlzoAAAAANDZNHZ8oDgBEII6rzAAAABb3h+EJBAAA"
	got, err := DecodeData(data)
	if err != nil {
		t.Errorf("DecodeData() error = %v", err)
		return
	}

	want := []byte{0x03, 0xc0, 0x63, 0x97, 0x3a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xc0, 0x63, 0x97, 0x3a, 0x00, 0x00, 0x00,
		0x00, 0x34, 0x36, 0x4d, 0x1d, 0x9f, 0x28, 0x0e, 0x00, 0x44, 0x20, 0x8e, 0xab, 0xcc, 0x00, 0x00,
		0x00, 0x16, 0xf7, 0x87, 0xe1, 0x09, 0x04, 0x00, 0x00}

	if !bytes.Equal(got, want) {
		t.Errorf("DecodeData() got = %v, want %v", got, want)
	}
}

func TestBinaryToSwapEvent(t *testing.T) {
	data := []byte{0x03, 0xc0, 0x63, 0x97, 0x3a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xc0, 0x63, 0x97, 0x3a, 0x00, 0x00, 0x00,
		0x00, 0x34, 0x36, 0x4d, 0x1d, 0x9f, 0x28, 0x0e, 0x00, 0x44, 0x20, 0x8e, 0xab, 0xcc, 0x00, 0x00,
		0x00, 0x16, 0xf7, 0x87, 0xe1, 0x09, 0x04, 0x00, 0x00}

	got, err := binaryToSwapEvent(data)
	if err != nil {
		t.Errorf("BinaryToTradeEvent() error = %v", err)
		return
	}

	want := EventData{
		LogType:          3,
		AmountIn:         983000000,
		MinimumAmountOut: 0,
		Direction:        1,
		UserSource:       983000000,
		PoolCoin:         3985313530459700,
		PoolPC:           879051546692,
		AmountOut:        4440485000982,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("BinaryToTradeEvent() got = %v, want %v", got, want)
	}
}

func TestBinaryToSwapEventBuy(t *testing.T) {
	//https://solana.fm/tx/vxE5kS73tHJFQHhY65HAtH6uvdzKUq4itJCD6rN5ih12dcrUYQAeBRQKDLdjwvcvQivtkgHwiaYpd8DemKXT7eQ
	data, err := DecodeData("AwAQXl8AAAAA0WPMcQ4AAAABAAAAAAAAAAAQXl8AAAAAUgMYKNYRAABb6qfEOAAAAPUG9LEdAAAA")
	if err != nil {
		t.Errorf("DecodeData() error = %v", err)
		return
	}

	got, err := binaryToSwapEvent(data)
	if err != nil {
		t.Errorf("BinaryToTradeEvent() error = %v", err)
		return
	}

	want := EventData{
		LogType:          3,
		AmountIn:         1600000000,
		MinimumAmountOut: 62038762449,
		Direction:        1,
		UserSource:       1600000000,
		PoolCoin:         19611493335890,
		PoolPC:           243817507419,
		AmountOut:        127539611381,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("BinaryToTradeEvent() got = %v, want %v", got, want)
	}
}

func TestBinaryToSwapEventSell(t *testing.T) {
	//https://solana.fm/tx/2tP1Z21Rm6ccWyxxsw2YaaWVmyvEHC2cBi3mTm7JqwuTzUjDDyrZuKXNrpdG3muvw66PcdCdTs4Ajodgs7TAawNz
	data, err := DecodeData("A+ClmPsNAAAAW42cEAAAAAACAAAAAAAAAJwMnaIWAAAA3qQ4uUwUAACNJCqdLQAAANvaQR8AAAAA")

	got, err := binaryToSwapEvent(data)
	if err != nil {
		t.Errorf("BinaryToTradeEvent() error = %v", err)
		return
	}

	want := EventData{
		LogType:          3,
		AmountIn:         60055660000,
		MinimumAmountOut: 278695259,
		Direction:        2,
		UserSource:       97217481884,
		PoolCoin:         22319757567198,
		PoolPC:           195910313101,
		AmountOut:        524409563,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("BinaryToTradeEvent() got = %v, want %v", got, want)
	}
}

func TestBinaryToSwapData(t *testing.T) {
	data := []byte{0x09, 0xc0, 0x63, 0x97, 0x3a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	sData, err := binaryToSwapData(data)
	if err != nil {
		t.Errorf("binaryToSwapData() error = %v", err)
		return
	}

	want := SwapData{
		Discriminator:    9,
		AmountIn:         983000000,
		MinimumAmountOut: 0,
	}

	if !reflect.DeepEqual(sData, want) {
		t.Errorf("binaryToSwapData() got = %v, want %v", sData, want)
	}
}
