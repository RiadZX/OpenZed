package utils

import "github.com/gagliardetto/solana-go"

func GetATA(wallet string, mint string) (string, error) {
	ata, _, err := solana.FindAssociatedTokenAddress(
		solana.MustPublicKeyFromBase58(wallet),
		solana.MustPublicKeyFromBase58(mint),
	)
	if err != nil {
		return "", err
	}
	return ata.String(), nil
}
