package raydiumsdk

import (
	"Zed/models"
	"Zed/raydiumsdk/ray_idl"
	"Zed/utils"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gookit/slog"
	"strconv"
)

// GetSwapInstruction returns the swap instruction from a transaction
func GetSwapInstruction(client models.IClient, transaction *solana.Transaction, innerInstructions []rpc.InnerInstruction) (ray_idl.SwapBaseIn, error) {
	if transaction == nil {
		return ray_idl.SwapBaseIn{}, fmt.Errorf("transaction is nil")
	}
	for _, i := range transaction.Message.Instructions {

		if len(i.Accounts) == 18 {
			swapIx, err := swapInstructionFromCompiledInstruction(client, i, transaction)
			if err != nil {
				slog.Debug("Error getting swap instruction from compiled instruction: " + err.Error())
				continue
			}
			return swapIx, nil
		}
	}

	//Not found check in inner instructions
	for _, innerInstruction := range innerInstructions {
		for _, i := range innerInstruction.Instructions {
			if len(i.Accounts) == 18 {
				swapIx, err := swapInstructionFromCompiledInstruction(client, i, transaction)
				if err != nil {
					slog.Debug("Error getting swap instruction from compiled instruction: " + err.Error())
					continue
				}
				if swapIx.GetSerumProgramAccount().PublicKey.String() != "srmqPvymJeFKQ4zGQed1GFppgkRHL9kaELCbyksJtPX" {
					slog.Debug("Serum program account is invalid: " + swapIx.GetSerumProgramAccount().PublicKey.String())
					continue
				}
				return swapIx, nil
			}
		}
	}

	//log the data to help with debugging
	slog.Debug("Instructions")
	for _, i := range transaction.Message.Instructions {
		slog.Debug(i)
		slog.Debug("Length of accounts: " + strconv.Itoa(len(i.Accounts)))
	}
	slog.Debug("Inner instructions")
	for _, innerInstruction := range innerInstructions {
		for _, i := range innerInstruction.Instructions {
			slog.Debug(i)
			slog.Debug("Length of accounts: " + strconv.Itoa(len(i.Accounts)))
		}
	}
	//Check in inner instructions
	return ray_idl.SwapBaseIn{}, fmt.Errorf("swap instruction not found")
}

func swapInstructionFromCompiledInstruction(client models.IClient, i solana.CompiledInstruction, transaction *solana.Transaction) (swapix ray_idl.SwapBaseIn, err error) {
	//This is most likely a swap instruction
	data := i.Data.String()
	decodedData, err := DecodeData(data)
	if err != nil {
		slog.Error("Error decoding data: " + err.Error())
		return ray_idl.SwapBaseIn{}, err
	}

	sData, err := binaryToSwapData(decodedData)
	if err != nil {
		slog.Error("Error converting binary to swap data: " + err.Error())
		return ray_idl.SwapBaseIn{}, err
	}
	sig := transaction.Signatures[0].String()

	transaction, err = utils.ResolveAddressTables(client, transaction)
	if err != nil {
		slog.Error("Error resolving address tables: " + err.Error() + " Signature: " + sig)
		return ray_idl.SwapBaseIn{}, err
	}

	instAccounts, err := i.ResolveInstructionAccounts(&transaction.Message)
	if err != nil {
		slog.Error("Error resolving instruction accounts: " + err.Error())
		return ray_idl.SwapBaseIn{}, err
	}

	swapInstruction, err := createSwapInstructionFromAccountsAndData(instAccounts, sData)
	if err != nil {
		slog.Error("Error creating swap instruction: " + err.Error())
		return ray_idl.SwapBaseIn{}, err
	}

	return swapInstruction, nil
}

func createSwapInstructionFromAccountsAndData(instAccounts []*solana.AccountMeta, swapData SwapData) (ray_idl.SwapBaseIn, error) {
	var swapInstruction *ray_idl.SwapBaseIn
	if len(instAccounts) == 18 {
		swapInstruction = ray_idl.NewSwapBaseInInstruction(
			swapData.AmountIn,
			swapData.MinimumAmountOut,
			instAccounts[0].PublicKey,  // tokenProgram
			instAccounts[1].PublicKey,  // amm
			instAccounts[2].PublicKey,  // ammAuthority
			instAccounts[3].PublicKey,  // ammOpenOrders
			instAccounts[4].PublicKey,  // ammTargetOrders
			instAccounts[5].PublicKey,  // poolCoinTokenAccount
			instAccounts[6].PublicKey,  // poolPcTokenAccount
			instAccounts[7].PublicKey,  // serumProgram
			instAccounts[8].PublicKey,  // serumMarket
			instAccounts[9].PublicKey,  // serumBids
			instAccounts[10].PublicKey, // serumAsks
			instAccounts[11].PublicKey, // serumEventQueue
			instAccounts[12].PublicKey, // serumCoinVaultAccount
			instAccounts[13].PublicKey, // serumPcVaultAccount
			instAccounts[14].PublicKey, // serumVaultSigner
			instAccounts[15].PublicKey, // userSourceTokenAccount
			instAccounts[16].PublicKey, // userDestinationTokenAccount
			instAccounts[17].PublicKey, // userSourceOwner
		)
	}
	if swapInstruction == nil {
		return ray_idl.SwapBaseIn{}, fmt.Errorf("swap instruction is nil")
	}
	if swapInstruction.Validate() != nil {
		return ray_idl.SwapBaseIn{}, fmt.Errorf("swap instruction is invalid")
	}

	return *swapInstruction, nil
}
