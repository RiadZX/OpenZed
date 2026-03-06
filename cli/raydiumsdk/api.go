package raydiumsdk

import (
	"Zed/models"
	"Zed/utils"
	"context"
	"encoding/binary"
	"errors"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"math"
	"strconv"
	"time"
)

// GetPricePerToken returns the price per token in lamports. buy = true -> buy tx, buy = false -> sell tx
func GetPricePerToken(client models.IClient, address string, delay int64, retries int64) (pricePerToken uint64, baseIsSol bool, mint solana.PublicKey, decimalsMint int, supply int, err error) {
	if retries == 0 {
		return 0, false, solana.PublicKey{}, 0, 0, errors.New("retries exhausted for adress: " + address)
	}
	if client == nil || address == "" {
		return 0, false, solana.PublicKey{}, 0, 0, errors.New("client or address is nil")
	}

	out, err := client.GetAccountInfoWithOpts(context.TODO(),
		solana.MustPublicKeyFromBase58(address),
		&rpc.GetAccountInfoOpts{
			Commitment: rpc.CommitmentProcessed,
		},
	)
	if err != nil {
		time.Sleep(time.Duration(delay) * time.Millisecond)
		return GetPricePerToken(client, address, delay, retries-1)
	}
	if out.Value == nil {
		return 0, false, solana.PublicKey{}, 0, 0, errors.New("account not found")
	}
	if out.Value.Data == nil {
		return 0, false, solana.PublicKey{}, 0, 0, errors.New("account data not found")
	}
	data := out.Value.Data.GetBinary()
	if len(data) == 0 {
		return 0, false, solana.PublicKey{}, 0, 0, errors.New("account data is empty")
	}

	//Get the data using offsets and lengths.
	//Offsets are in bytes.
	baseDecimalOffset := 4
	baseDecimalLength := 8
	var baseDecimal = binary.LittleEndian.Uint64(data[baseDecimalOffset*8 : baseDecimalOffset*8+baseDecimalLength*8])

	quoteDecimalOffset := 5
	quoteDecimalLength := 8
	var quoteDecimal = binary.LittleEndian.Uint64(data[quoteDecimalOffset*8 : quoteDecimalOffset*8+quoteDecimalLength*8])

	baseVaultOffset := 42
	baseVaultLength := 4
	var baseVault = solana.PublicKeyFromBytes(data[baseVaultOffset*8 : baseVaultOffset*8+baseVaultLength*8])

	quoteVaultOffset := 46
	quoteVaultLength := 4
	var quoteVault = solana.PublicKeyFromBytes(data[quoteVaultOffset*8 : quoteVaultOffset*8+quoteVaultLength*8])

	var baseMintOffset = 50
	var baseMintLength = 4
	var baseMint = solana.PublicKeyFromBytes(data[baseMintOffset*8 : baseMintOffset*8+baseMintLength*8])

	var quoteMintOffset = 54
	var quoteMintLength = 4
	var quoteMint = solana.PublicKeyFromBytes(data[quoteMintOffset*8 : quoteMintOffset*8+quoteMintLength*8])

	//Get the tokenbalances for baseVault and quotevault
	//price = (quoteBalance/quoteDecimal) / (baseBalance/baseDecimal)
	quoteTokenBalance, err := utils.GetTokenBalance(client, quoteVault, delay, retries)
	if err != nil {
		return 0, false, solana.PublicKey{}, 0, 0, err
	}
	baseTokenBalance, err := utils.GetTokenBalance(client, baseVault, delay, retries)
	if err != nil {
		return 0, false, solana.PublicKey{}, 0, 0, err
	}

	quoteBalance, err := strconv.Atoi(quoteTokenBalance.Amount)
	if err != nil {
		return 0, false, solana.PublicKey{}, 0, 0, err
	}
	baseBalance, err := strconv.Atoi(baseTokenBalance.Amount)
	if err != nil {
		return 0, false, solana.PublicKey{}, 0, 0, err
	}
	if baseMint.String() == Wsol.String() {
		return uint64((float64(baseBalance)) / (float64(quoteBalance)) * math.Pow(10, float64(quoteDecimal))), true, quoteMint, int(quoteDecimal), quoteBalance, nil
	}
	if quoteMint.String() == Wsol.String() {
		return uint64((float64(quoteBalance)) / (float64(baseBalance)) * math.Pow(10, float64(baseDecimal))), false, baseMint, int(baseDecimal), baseBalance, nil
	}

	return 0, false, solana.PublicKey{}, 0, 0, errors.New("neither base nor quote mint is wsol")
}
