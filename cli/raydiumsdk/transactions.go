package raydiumsdk

import (
	"Zed/raydiumsdk/ray_idl"
	"bytes"
	"fmt"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/scribesavant/raydium-swap-go/raydium/constants"
)

type swapV4Instruction struct {
	Instruction      uint8
	AmountIn         uint64
	MinimumOutAmount uint64
}

func (inst *swapV4Instruction) Data() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := bin.NewBorshEncoder(buf).Encode(inst); err != nil {
		return nil, fmt.Errorf("unable to encode instruction: %w", err)
	}
	return buf.Bytes(), nil
}

func NewSwapV4Instruction(
	copySwapInstruction *ray_idl.SwapBaseIn,
	amountIn uint64,
	minAmountOut uint64,
	tokenAccountIn solana.PublicKey,
	tokenAccountOut solana.PublicKey,
	signer solana.PublicKey) (solana.Instruction, error) {
	inst := &swapV4Instruction{
		Instruction:      9,
		AmountIn:         amountIn,
		MinimumOutAmount: minAmountOut,
	}
	keys := solana.AccountMetaSlice{}

	keys.Append(solana.Meta(solana.TokenProgramID))
	keys.Append(copySwapInstruction.GetAmmAccount())
	keys.Append(copySwapInstruction.GetAmmAuthorityAccount())
	keys.Append(copySwapInstruction.GetAmmOpenOrdersAccount())

	//if poolKeys.Version == 4 { ASSUME ALWAYS 4
	keys.Append(copySwapInstruction.GetAmmTargetOrdersAccount())
	//}

	keys.Append(copySwapInstruction.GetPoolCoinTokenAccountAccount())
	keys.Append(copySwapInstruction.GetPoolPcTokenAccountAccount())

	//if poolKeys.Version == 5 {
	//	keys.Append(solana.Meta(constants.ModelDataPubkey).WRITE())
	//}
	keys.Append(copySwapInstruction.GetSerumProgramAccount())
	keys.Append(copySwapInstruction.GetSerumMarketAccount())
	keys.Append(copySwapInstruction.GetSerumBidsAccount())
	keys.Append(copySwapInstruction.GetSerumAsksAccount())
	keys.Append(copySwapInstruction.GetSerumEventQueueAccount())
	keys.Append(copySwapInstruction.GetSerumCoinVaultAccountAccount())
	keys.Append(copySwapInstruction.GetSerumPcVaultAccountAccount())
	keys.Append(copySwapInstruction.GetSerumVaultSignerAccount())
	keys.Append(solana.Meta(tokenAccountIn).WRITE())
	keys.Append(solana.Meta(tokenAccountOut).WRITE())
	keys.Append(solana.Meta(signer).SIGNER())

	data, err := inst.Data()
	if err != nil {
		return nil, err
	}
	intstr := solana.NewInstruction(
		constants.RAYDIUM_V4_PROGRAM_ID,
		keys,
		data,
	)

	return intstr, nil
}
