package utils

import (
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gookit/slog"
	geyser_pb "github.com/weeaa/goyser/pb"
	"strconv"
)

func ParseInnerInstructions(instructions []*geyser_pb.InnerInstructions) []rpc.InnerInstruction {
	var innerInstructions = make([]rpc.InnerInstruction, len(instructions))
	for _, i := range instructions {
		var compiledIxs = make([]solana.CompiledInstruction, len(i.Instructions))
		for _, c := range i.Instructions {
			//convert the accounts from []uint16 to []byte
			compiledIxs = append(compiledIxs, solana.CompiledInstruction{
				ProgramIDIndex: uint16(c.ProgramIdIndex),
				Accounts:       ByteSliceToUint16Slice(c.Accounts),
				Data:           c.Data,
			})
		}
		innerInstructions = append(innerInstructions, rpc.InnerInstruction{
			Index:        uint16(i.Index),
			Instructions: compiledIxs,
		})
	}
	return innerInstructions
}

// CustomPrePostTokenBalances Data structure to hold and calculate the tokenbalances, for both ws and grpc
type CustomPrePostTokenBalances struct {
	PreTokenBalances  []XTokenBalance
	PostTokenBalances []XTokenBalance
}

func (c *CustomPrePostTokenBalances) GetDecimals(mint string) uint8 {
	for _, balance := range c.PreTokenBalances {
		if balance.Mint == mint {
			return balance.Decimals
		}
	}
	for _, balance := range c.PostTokenBalances {
		if balance.Mint == mint {
			return balance.Decimals
		}
	}
	return 0
}

// FromGeyserPrePro fills the data structure from geyser data
func (c *CustomPrePostTokenBalances) FromGeyserPrePro(preTokenBalances []*geyser_pb.TokenBalance, postTokenBalance []*geyser_pb.TokenBalance) error {
	for _, balance := range postTokenBalance {
		tokenites, err := strconv.ParseFloat(balance.UiTokenAmount.Amount, 64)
		if err != nil {
			slog.Error("Math error")
			return fmt.Errorf("math error")
		}
		ptb := XTokenBalance{
			Mint:         balance.Mint,
			Owner:        balance.Owner,
			Tokenites:    tokenites,
			Decimals:     uint8(balance.UiTokenAmount.Decimals),
			AccountIndex: uint16(balance.AccountIndex),
		}
		c.PostTokenBalances = append(c.PostTokenBalances, ptb)
	}
	for _, balance := range preTokenBalances {
		tokenites, err := strconv.ParseFloat(balance.UiTokenAmount.Amount, 64)
		if err != nil {
			slog.Error("Math error")
			return fmt.Errorf("math error")
		}
		ptb := XTokenBalance{
			Mint:         balance.Mint,
			Owner:        balance.Owner,
			Tokenites:    tokenites,
			Decimals:     uint8(balance.UiTokenAmount.Decimals),
			AccountIndex: uint16(balance.AccountIndex),
		}
		c.PreTokenBalances = append(c.PreTokenBalances, ptb)
	}
	return nil
}

func (c *CustomPrePostTokenBalances) FromSolanaPrePro(preTokenBalances []rpc.TokenBalance, postTokenBalance []rpc.TokenBalance) error {
	for _, balance := range postTokenBalance {
		tokenites, err := strconv.ParseFloat(balance.UiTokenAmount.Amount, 64)
		if err != nil {
			slog.Error("Math error")
		}
		ptb := XTokenBalance{
			AccountIndex: balance.AccountIndex,
			Mint:         balance.Mint.String(),
			Owner:        balance.Owner.String(),
			Decimals:     balance.UiTokenAmount.Decimals,
			Tokenites:    tokenites,
		}
		c.PostTokenBalances = append(c.PostTokenBalances, ptb)
	}
	for _, balance := range preTokenBalances {
		tokenites, err := strconv.ParseFloat(balance.UiTokenAmount.Amount, 64)
		if err != nil {
			slog.Error("Math error")
		}

		ptb := XTokenBalance{
			AccountIndex: balance.AccountIndex,
			Mint:         balance.Mint.String(),
			Owner:        balance.Owner.String(),
			Decimals:     balance.UiTokenAmount.Decimals,
			Tokenites:    tokenites,
		}
		c.PreTokenBalances = append(c.PreTokenBalances, ptb)
	}
	return nil
}

// GetPreTokenBalance Gets the pretoken balance by owner and mint
// if - -> not found,or had 0 balance
func (c *CustomPrePostTokenBalances) GetPreTokenBalance(owner, mint string) float64 {
	for i := 0; i < len(c.PreTokenBalances); i++ {
		x := c.PreTokenBalances[i]
		if x.Owner == owner && x.Mint == mint {
			return x.Tokenites
		}
	}
	return 0
}

// GetPostTokenBalance Gets the posttoken balance by owner and mint
// if 0 -> not found, or had 0 balance
func (c *CustomPrePostTokenBalances) GetPostTokenBalance(owner, mint string) float64 {
	for i := 0; i < len(c.PostTokenBalances); i++ {
		x := c.PostTokenBalances[i]
		if x.Owner == owner && x.Mint == mint {
			return x.Tokenites
		}
	}
	return 0
}

// XTokenBalance shared data structure between grpc and ws
type XTokenBalance struct {
	AccountIndex uint16
	Mint         string
	Owner        string
	Decimals     uint8
	UIAmount     int
	Tokenites    float64
}
