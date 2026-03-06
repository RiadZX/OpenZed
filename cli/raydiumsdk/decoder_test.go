package raydiumsdk

import (
	"Zed/utils"
	"github.com/gagliardetto/solana-go"
	"testing"
)

func Test_GetDecimals(t *testing.T) {
	pptb := &utils.CustomPrePostTokenBalances{
		PreTokenBalances: []utils.XTokenBalance{
			{Mint: "token1", Decimals: 6},
		},
		PostTokenBalances: []utils.XTokenBalance{
			{Mint: "token2", Decimals: 8},
		},
	}

	decimals := GetDecimals(pptb, "token1")
	if decimals == nil || *decimals != 6 {
		t.Errorf("GetDecimals() got = %v, want = 6", decimals)
	}

	decimals = GetDecimals(pptb, "token2")
	if decimals == nil || *decimals != 8 {
		t.Errorf("GetDecimals() got = %v, want = 8", decimals)
	}

	decimals = GetDecimals(pptb, "token3")
	if decimals != nil {
		t.Errorf("GetDecimals() got = %v, want = nil", decimals)
	}
}

func Test_GetSupply(t *testing.T) {

	pptb := &utils.CustomPrePostTokenBalances{
		PreTokenBalances: []utils.XTokenBalance{
			{AccountIndex: 0, Mint: "token1", Owner: rayAuthority.String(), Tokenites: 1000000000, Decimals: 9},
		},
		PostTokenBalances: []utils.XTokenBalance{
			{AccountIndex: 0, Mint: "token1", Owner: rayAuthority.String(), Tokenites: 2000000000, Decimals: 9},
		},
	}

	supply := GetSupply(pptb, true, "token1")
	if supply == nil || *supply != 2 {
		t.Errorf("GetSupply() got = %v, want = 2", supply)
	}

	supply = GetSupply(pptb, false, "token1")
	if supply == nil || *supply != 1 {
		t.Errorf("GetSupply() got = %v, want = 1", supply)
	}

	supply = GetSupply(pptb, true, "token2")
	if supply != nil {
		t.Errorf("GetSupply() got = %v, want = nil", supply)
	}
}

func Test_GetMintFromBalances(t *testing.T) {
	rayAuthority := solana.MustPublicKeyFromBase58("5Q544fKrFoe6tsEbD7S8EmxGTJYAKtTVhAW5Q5pge4j1")
	wsol := solana.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")
	signer := solana.MustPublicKeyFromBase58("BrNoqdHUCcv9yTncnZeSjSov8kqhpmzv1nAiPbq1M95H")

	pptb := &utils.CustomPrePostTokenBalances{
		PreTokenBalances: []utils.XTokenBalance{
			{AccountIndex: 0, Mint: wsol.String(), Owner: rayAuthority.String(), Tokenites: 10, Decimals: 9},
			{AccountIndex: 1, Mint: "token1", Owner: signer.String(), Tokenites: 5, Decimals: 9},
		},
		PostTokenBalances: []utils.XTokenBalance{
			{AccountIndex: 0, Mint: wsol.String(), Owner: rayAuthority.String(), Tokenites: 15, Decimals: 9},
			{AccountIndex: 1, Mint: "token1", Owner: signer.String(), Tokenites: 8, Decimals: 9},
		},
	}

	isBuy, token := GetMintFromBalances(pptb, signer)
	if !isBuy || token != "token1" {
		t.Errorf("GetMintFromBalances() got isBuy = %v, token = %v, want isBuy = true, token = token1", isBuy, token)
	}
}

func Test_GetMintFromBalances2(t *testing.T) {
	rayAuthority := solana.MustPublicKeyFromBase58("5Q544fKrFoe6tsEbD7S8EmxGTJYAKtTVhAW5Q5pge4j1")
	wsol := solana.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")
	signer := solana.MustPublicKeyFromBase58("BrNoqdHUCcv9yTncnZeSjSov8kqhpmzv1nAiPbq1M95H")

	pptb := &utils.CustomPrePostTokenBalances{
		PreTokenBalances: []utils.XTokenBalance{
			{AccountIndex: 0, Mint: wsol.String(), Owner: rayAuthority.String(), Tokenites: 15, Decimals: 9},
			{AccountIndex: 1, Mint: "token1", Owner: signer.String(), Tokenites: 8, Decimals: 9},
		},
		PostTokenBalances: []utils.XTokenBalance{
			{AccountIndex: 1, Mint: "token1", Owner: signer.String(), Tokenites: 5, Decimals: 9},
			{AccountIndex: 0, Mint: wsol.String(), Owner: rayAuthority.String(), Tokenites: 10, Decimals: 9},
		}}

	isBuy, token := GetMintFromBalances(pptb, signer)
	if isBuy || token != "token1" {
		t.Errorf("GetMintFromBalances() got isBuy = %v, token = %v, want isBuy = false, token = token1", isBuy, token)
	}
}

func Test_GetMintFromBalances3(t *testing.T) {
	rayAuthority := solana.MustPublicKeyFromBase58("5Q544fKrFoe6tsEbD7S8EmxGTJYAKtTVhAW5Q5pge4j1")
	wsol := solana.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")
	signer := solana.MustPublicKeyFromBase58("BrNoqdHUCcv9yTncnZeSjSov8kqhpmzv1nAiPbq1M95H")

	pptb := &utils.CustomPrePostTokenBalances{
		PreTokenBalances: []utils.XTokenBalance{
			{AccountIndex: 0, Mint: wsol.String(), Owner: rayAuthority.String(), Tokenites: 10, Decimals: 9},
			{AccountIndex: 1, Mint: "token1", Owner: signer.String(), Tokenites: 5, Decimals: 9},
		},
		PostTokenBalances: []utils.XTokenBalance{
			{AccountIndex: 0, Mint: wsol.String(), Owner: rayAuthority.String(), Tokenites: 15, Decimals: 9},
			{AccountIndex: 2, Mint: "token2", Owner: signer.String(), Tokenites: 8, Decimals: 9},
		},
	}

	isBuy, token := GetMintFromBalances(pptb, signer)
	if !isBuy || token != "token2" {
		t.Errorf("GetMintFromBalances() got isBuy = %v, token = %v, want isBuy = true, token = token1", isBuy, token)
	}
}
