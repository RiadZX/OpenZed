package mocks

import (
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/mr-tron/base58"
)

// Fake data for testing purposes

// ----------------------------------------------
var pFSwapCustomProgramTransaction = solana.Transaction{
	Signatures: nil,
	Message: solana.Message{
		AccountKeys:         nil,
		Header:              solana.MessageHeader{},
		RecentBlockhash:     solana.Hash{},
		Instructions:        []solana.CompiledInstruction{randomIX},
		AddressTableLookups: nil,
	},
}
var randomIX = solana.CompiledInstruction{
	ProgramIDIndex: 0,
	Accounts:       []uint16{0, 1, 2, 3, 4, 5, 6, 7},
	Data:           []byte{0x01, 0x02, 0x03, 0x04},
}
var pfSwapixData, _ = base58.Decode("AJTQ2h9DXrBdC1Fdc6W7Dq4v6KjBxNdLj")
var pfSwapix = solana.CompiledInstruction{
	ProgramIDIndex: 0,
	Accounts:       []uint16{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	Data:           pfSwapixData,
}

func GetPFSwapCustomProgramBuyTransaction() (solana.Transaction, []rpc.InnerInstruction) {
	//init the account keys
	var accountKeys solana.PublicKeySlice
	accountKeys = append(accountKeys, solana.MustPublicKeyFromBase58("CV9aBYXNADXwGQiUM1Rmr8hrSLKLiam4u7DY3UEcHCA4"))
	accountKeys = append(accountKeys, solana.MustPublicKeyFromBase58("BV9aBYXNADXwGQiUM1Rmr8hrSLKLiam4u7DY3UEcHCA4"))
	accountKeys = append(accountKeys, solana.MustPublicKeyFromBase58("AV9aBYXNADXwGQiUM1Rmr8hrSLKLiam4u7DY3UEcHCA4"))
	accountKeys = append(accountKeys, solana.MustPublicKeyFromBase58("rE3JG5gUXzGUvqSX2CvydsAzm1bZ4FsFgXocpyYXRA6"))
	accountKeys = append(accountKeys, solana.MustPublicKeyFromBase58("HV9aBYXNADXwGQiUM1Rmr8hrSLKLiam4u7DY3UEcHCA4"))

	pFSwapCustomProgramTransaction.Message.AccountKeys = accountKeys

	//inner instructions
	return pFSwapCustomProgramTransaction, []rpc.InnerInstruction{{
		Index:        0,
		Instructions: []solana.CompiledInstruction{pfSwapix},
	}}
}

// ------------------------------------------------------------
var pfSwapixDataSell, _ = base58.Decode("5jRcjdixRUDVzKrjnjHs65TH8RGdTJ4PH")
var pfSwapixSell = solana.CompiledInstruction{
	ProgramIDIndex: 0,
	Accounts:       []uint16{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	Data:           pfSwapixDataSell,
}

func GetPFSwapCustomProgramSellTransaction() (solana.Transaction, []rpc.InnerInstruction) {
	//init the account keys
	var accountKeys solana.PublicKeySlice
	accountKeys = append(accountKeys, solana.MustPublicKeyFromBase58("CV9aBYXNADXwGQiUM1Rmr8hrSLKLiam4u7DY3UEcHCA4"))
	accountKeys = append(accountKeys, solana.MustPublicKeyFromBase58("BV9aBYXNADXwGQiUM1Rmr8hrSLKLiam4u7DY3UEcHCA4"))
	accountKeys = append(accountKeys, solana.MustPublicKeyFromBase58("AV9aBYXNADXwGQiUM1Rmr8hrSLKLiam4u7DY3UEcHCA4"))
	accountKeys = append(accountKeys, solana.MustPublicKeyFromBase58("rE3JG5gUXzGUvqSX2CvydsAzm1bZ4FsFgXocpyYXRA6"))
	accountKeys = append(accountKeys, solana.MustPublicKeyFromBase58("HV9aBYXNADXwGQiUM1Rmr8hrSLKLiam4u7DY3UEcHCA4"))

	pFSwapCustomProgramTransaction.Message.AccountKeys = accountKeys

	//inner instructions

	return pFSwapCustomProgramTransaction, []rpc.InnerInstruction{{
		Index:        0,
		Instructions: []solana.CompiledInstruction{pfSwapixSell},
	}}
}

// ------------------------------------------------------------
var pFSwapProgramTransaction = solana.Transaction{
	Signatures: nil,
	Message: solana.Message{
		AccountKeys:         nil,
		Header:              solana.MessageHeader{},
		RecentBlockhash:     solana.Hash{},
		Instructions:        []solana.CompiledInstruction{pfSwapix},
		AddressTableLookups: nil,
	},
}

func GetPFSwapBuyTransaction() (solana.Transaction, []rpc.InnerInstruction) {
	//init the account keys
	var accountKeys solana.PublicKeySlice
	accountKeys = append(accountKeys, solana.MustPublicKeyFromBase58("CV9aBYXNADXwGQiUM1Rmr8hrSLKLiam4u7DY3UEcHCA4"))
	accountKeys = append(accountKeys, solana.MustPublicKeyFromBase58("BV9aBYXNADXwGQiUM1Rmr8hrSLKLiam4u7DY3UEcHCA4"))
	accountKeys = append(accountKeys, solana.MustPublicKeyFromBase58("AV9aBYXNADXwGQiUM1Rmr8hrSLKLiam4u7DY3UEcHCA4"))
	accountKeys = append(accountKeys, solana.MustPublicKeyFromBase58("rE3JG5gUXzGUvqSX2CvydsAzm1bZ4FsFgXocpyYXRA6"))
	accountKeys = append(accountKeys, solana.MustPublicKeyFromBase58("HV9aBYXNADXwGQiUM1Rmr8hrSLKLiam4u7DY3UEcHCA4"))

	pFSwapProgramTransaction.Message.AccountKeys = accountKeys
	return pFSwapProgramTransaction, []rpc.InnerInstruction{}
}

// ------------------------------------------------------------
