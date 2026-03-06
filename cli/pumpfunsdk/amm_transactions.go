package pumpfunsdk

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gookit/slog"
	"github.com/mr-tron/base58"
)

type amm_buy_instruction_data struct {
	BaseAmountOut    uint64
	MaxQuoteAmountIn uint64
}

func (inst *amm_buy_instruction_data) Data() ([]byte, error) {
	// Create a buffer to hold the discriminator and instruction data
	buf := new(bytes.Buffer)

	// Write the 8-byte instruction discriminator
	// This should match the discriminator in the Rust/Anchor program
	// You'll need to replace this with the actual discriminator for the "buy" instruction
	discriminator := []byte{102, 6, 61, 18, 1, 218, 235, 234} // Placeholder - replace with correct discriminator

	// Write the discriminator first
	if _, err := buf.Write(discriminator); err != nil {
		return nil, fmt.Errorf("failed to write discriminator: %w", err)
	}

	// Then encode the rest of the instruction data
	if err := bin.NewBorshEncoder(buf).Encode(inst); err != nil {
		return nil, fmt.Errorf("unable to encode instruction: %w", err)
	}

	return buf.Bytes(), nil
}

func NewAMMBuyInstruction(
	max_quote_amount_in uint64,
	base_token_out uint64,
	amm_buy_event AMMBuyEvent,
	amm_buy_tx AMMBuyTransaction,
	user_base_token_account solana.PublicKey,
	user_quote_token_account solana.PublicKey,
	protocol_fee_recipient solana.PublicKey,
	protocol_fee_recipient_token_account solana.PublicKey,
	signer solana.PublicKey,
) (solana.Instruction, error) {
	// we need to create an instruction, set the data and set the accounts for the instruction
	inst := &amm_buy_instruction_data{
		BaseAmountOut:    base_token_out,
		MaxQuoteAmountIn: max_quote_amount_in,
	}
	keys := solana.AccountMetaSlice{}

	// #1 - Pool:
	keys.Append(solana.Meta(amm_buy_event.Pool))
	// #2 - User:
	keys.Append(solana.Meta(signer).SIGNER().WRITE())
	// #3 - Global Config:
	keys.Append(solana.Meta(amm_buy_tx.GlobalConfig))
	// #4 - Base Mint:
	keys.Append(solana.Meta(amm_buy_tx.BaseMint))
	// #5 - Quote Mint:
	keys.Append(solana.Meta(amm_buy_tx.QuoteMint))
	// #6 - User Base Token Account:
	keys.Append(solana.Meta(user_base_token_account).WRITE())
	// #7 - User Quote Token Account:
	keys.Append(solana.Meta(user_quote_token_account).WRITE())
	// #8 - Pool Base Token Account:
	keys.Append(solana.Meta(amm_buy_tx.PoolBaseTokenAccount).WRITE())
	// #9 - Pool Quote Token Account:
	keys.Append(solana.Meta(amm_buy_tx.PoolQuoteTokenAccount).WRITE())
	// #10 - Protocol Fee Recipient:
	keys.Append(solana.Meta(protocol_fee_recipient))
	// #11 - Protocol Fee Recipient Token Account:
	keys.Append(solana.Meta(protocol_fee_recipient_token_account).WRITE())
	// #12 - Base Token Program:
	keys.Append(solana.Meta(solana.TokenProgramID))
	// #13 - Quote Token Program:
	keys.Append(solana.Meta(solana.TokenProgramID))
	// #14 - System Program:
	keys.Append(solana.Meta(solana.SystemProgramID))
	// #15 - Associated Token Program:
	keys.Append(solana.Meta(solana.SPLAssociatedTokenAccountProgramID))
	// #16 - Event Authority:
	keys.Append(solana.Meta(solana.MustPublicKeyFromBase58("GS4CU59F31iL7aR2Q8zVS8DRrcRnXX1yjQ66TqNVQnaR")))
	// #17 - Program:
	keys.Append(solana.Meta(PUMP_AMM))
	// #18 - Coin Creator Vault ATA:
	keys.Append(solana.Meta(amm_buy_tx.CoinCreatorVaultATA).WRITE())
	// #19 - Coin Creator Vault Authority:
	keys.Append(solana.Meta(amm_buy_tx.CoinCreatorVaultAuthority).WRITE())

	data, err := inst.Data()
	if err != nil {
		return nil, err
	}
	intstr := solana.NewInstruction(
		PUMP_AMM,
		keys,
		data,
	)

	return intstr, nil
}

type amm_sell_instruction_data struct {
	BaseAmountIn     uint64
	MinQuoteAmountIn uint64
}

func (inst *amm_sell_instruction_data) Data() ([]byte, error) {
	// Create a buffer to hold the discriminator and instruction data
	buf := new(bytes.Buffer)

	// Write the 8-byte instruction discriminator
	// This should match the discriminator in the Rust/Anchor program
	// You'll need to replace this with the actual discriminator for the "sell" instruction
	discriminator := []byte{51, 230, 133, 164, 1, 127, 131, 173} // Placeholder - replace with correct discriminator

	// Write the discriminator first
	if _, err := buf.Write(discriminator); err != nil {
		return nil, fmt.Errorf("failed to write discriminator: %w", err)
	}

	// Then encode the rest of the instruction data
	if err := bin.NewBorshEncoder(buf).Encode(inst); err != nil {
		return nil, fmt.Errorf("unable to encode instruction: %w", err)
	}

	return buf.Bytes(), nil
}

func NewAMMSellInstruction(
	base_amount_in uint64,
	min_quote_amount_in uint64,
	amm_sell_event AMMSellEvent,
	amm_sell_tx AMMSellTransaction,
	user_base_token_account solana.PublicKey,
	user_quote_token_account solana.PublicKey,
	protocol_fee_recipient solana.PublicKey,
	protocol_fee_recipient_token_account solana.PublicKey,
	signer solana.PublicKey,
) (solana.Instruction, error) {
	// we need to create an instruction, set the data and set the accounts for the instruction
	inst := &amm_sell_instruction_data{
		BaseAmountIn:     base_amount_in,
		MinQuoteAmountIn: min_quote_amount_in,
	}
	keys := solana.AccountMetaSlice{}

	// #1 - Pool:
	keys.Append(solana.Meta(amm_sell_event.Pool))
	// #2 - User:
	keys.Append(solana.Meta(signer).SIGNER().WRITE())
	// #3 - Global Config:
	keys.Append(solana.Meta(amm_sell_tx.GlobalConfig))
	// #4 - Base Mint:
	keys.Append(solana.Meta(amm_sell_tx.BaseMint))
	// #5 - Quote Mint:
	keys.Append(solana.Meta(amm_sell_tx.QuoteMint))
	// #6 - User Base Token Account:
	keys.Append(solana.Meta(user_base_token_account).WRITE())
	// #7 - User Quote Token Account:
	keys.Append(solana.Meta(user_quote_token_account).WRITE())
	// #8 - Pool Base Token Account:
	keys.Append(solana.Meta(amm_sell_tx.PoolBaseTokenAccount).WRITE())
	// #9 - Pool Quote Token Account:
	keys.Append(solana.Meta(amm_sell_tx.PoolQuoteTokenAccount).WRITE())
	// #10 - Protocol Fee Recipient:
	keys.Append(solana.Meta(protocol_fee_recipient))
	// #11 - Protocol Fee Recipient Token Account:
	keys.Append(solana.Meta(protocol_fee_recipient_token_account).WRITE())
	// #12 - Base Token Program:
	keys.Append(solana.Meta(solana.TokenProgramID))
	// #13 - Quote Token Program:
	keys.Append(solana.Meta(solana.TokenProgramID))
	// #14 - System Program:
	keys.Append(solana.Meta(solana.SystemProgramID))
	// #15 - Associated Token Program:
	keys.Append(solana.Meta(solana.SPLAssociatedTokenAccountProgramID))
	// #16 - Event Authority:
	keys.Append(solana.Meta(solana.MustPublicKeyFromBase58("GS4CU59F31iL7aR2Q8zVS8DRrcRnXX1yjQ66TqNVQnaR")))
	// #17 - Program:
	keys.Append(solana.Meta(PUMP_AMM))
	// #18 - Coin Creator Vault ATA:
	keys.Append(solana.Meta(amm_sell_tx.CoinCreatorVaultATA).WRITE())
	// #19 - Coin Creator Vault Authority:
	keys.Append(solana.Meta(amm_sell_tx.CoinCreatorVaultAuthority).WRITE())

	data, err := inst.Data()
	if err != nil {
		return nil, err
	}
	intstr := solana.NewInstruction(
		PUMP_AMM,
		keys,
		data,
	)

	return intstr, nil
}

type AMMBuyTransaction struct {
	BaseAmountOut    uint64
	MaxQuoteAmountIn uint64

	Pool         solana.PublicKey
	User         solana.PublicKey
	GlobalConfig solana.PublicKey

	BaseMint  solana.PublicKey
	QuoteMint solana.PublicKey

	UserBaseTokenAccount  solana.PublicKey
	UserQuoteTokenAccount solana.PublicKey
	PoolBaseTokenAccount  solana.PublicKey
	PoolQuoteTokenAccount solana.PublicKey

	ProtocolFeeRecipient             solana.PublicKey
	ProtocolFeeRecipientTokenAccount solana.PublicKey

	CoinCreatorVaultATA       solana.PublicKey
	CoinCreatorVaultAuthority solana.PublicKey
}

func (b *AMMBuyTransaction) String() string {
	return "🛒 AMMBuyTransaction{" +
		" \n 📈 BaseAmountOut: " + string(b.BaseAmountOut) +
		",\n 💰 MaxQuoteAmountIn: " + string(b.MaxQuoteAmountIn) +
		",\n 🏊 Pool: " + b.Pool.String() +
		",\n 👤 User: " + b.User.String() +
		",\n 🌐 GlobalConfig: " + b.GlobalConfig.String() +
		",\n 🪙 BaseMint: " + b.BaseMint.String() +
		",\n 💱 QuoteMint: " + b.QuoteMint.String() +
		",\n 👛 UserBaseTokenAccount: " + b.UserBaseTokenAccount.String() +
		",\n 👛 UserQuoteTokenAccount: " + b.UserQuoteTokenAccount.String() +
		",\n 🏦 PoolBaseTokenAccount: " + b.PoolBaseTokenAccount.String() +
		",\n 🏦 PoolQuoteTokenAccount: " + b.PoolQuoteTokenAccount.String() +
		",\n 💸 ProtocolFeeReceipent: " + b.ProtocolFeeRecipient.String() +
		",\n 💸 ProtocolFeeReceipentTokenAccount: " + b.ProtocolFeeRecipientTokenAccount.String() +
		"}"
}

type AMMSellTransaction struct {
	BaseAmountIn      uint64
	MinQuoteAmountOut uint64

	Pool         solana.PublicKey
	User         solana.PublicKey
	GlobalConfig solana.PublicKey

	BaseMint  solana.PublicKey
	QuoteMint solana.PublicKey

	UserBaseTokenAccount  solana.PublicKey
	UserQuoteTokenAccount solana.PublicKey
	PoolBaseTokenAccount  solana.PublicKey
	PoolQuoteTokenAccount solana.PublicKey

	ProtocolFeeRecipient             solana.PublicKey
	ProtocolFeeRecipientTokenAccount solana.PublicKey

	CoinCreatorVaultATA       solana.PublicKey
	CoinCreatorVaultAuthority solana.PublicKey
}

func (s *AMMSellTransaction) String() string {
	return "🛒 AMMSellTransaction{" +
		" \n 📉 BaseAmountIn: " + string(s.BaseAmountIn) +
		",\n 💰 MinQuoteAmountOut: " + string(s.MinQuoteAmountOut) +
		",\n 🏊 Pool: " + s.Pool.String() +
		",\n 👤 User: " + s.User.String() +
		",\n 🌐 GlobalConfig: " + s.GlobalConfig.String() +
		",\n 🪙 BaseMint: " + s.BaseMint.String() +
		",\n 💱 QuoteMint: " + s.QuoteMint.String() +
		",\n 👛 UserBaseTokenAccount: " + s.UserBaseTokenAccount.String() +
		",\n 👛 UserQuoteTokenAccount: " + s.UserQuoteTokenAccount.String() +
		",\n 🏦 PoolBaseTokenAccount: " + s.PoolBaseTokenAccount.String() +
		",\n 🏦 PoolQuoteTokenAccount: " + s.PoolQuoteTokenAccount.String() +
		",\n 💸 ProtocolFeeRecipient: " + s.ProtocolFeeRecipient.String() +
		",\n 💸 ProtocolFeeRecipientTokenAccount: " + s.ProtocolFeeRecipientTokenAccount.String() +
		"}"
}

type AMMSwapTransaction struct {
	BuyTransaction  *AMMBuyTransaction
	SellTransaction *AMMSellTransaction
}

var (
	// account key indices
	poolIndex                             = 0
	userIndex                             = 1
	globalConfigIndex                     = 2
	baseMintIndex                         = 3
	quoteMintIndex                        = 4
	userBaseTokenIndex                    = 5
	userQuoteTokenIndex                   = 6
	poolBaseTokenIndex                    = 7
	poolQuoteTokenIndex                   = 8
	protocolFeeRecipientIndex             = 9
	protocolFeeRecipientTokenAccountIndex = 10
	coinCreatorVaultATAIndex              = 17
	coinCreatorVaultAuthorityIndex        = 18
)

func ParseAMMTransaction(tx *solana.Transaction, innerInstructions *[]rpc.InnerInstruction) (*AMMSwapTransaction, error) {
	if tx == nil || innerInstructions == nil {
		return nil, nil
	}

	// Look for the Buy or Sell transaction.
	var buyIdentifier = []byte{102, 6, 61, 18, 1, 218, 235, 234}
	var sellIdentifier = []byte{51, 230, 133, 164, 1, 127, 131, 173}

	var buyTransaction *AMMBuyTransaction
	var sellTransaction *AMMSellTransaction

	// Resolve lookups

	for _, i := range tx.Message.Instructions {
		decodedData, err := base58.Decode(i.Data.String())
		if err != nil {
			continue
		}
		if len(decodedData) < 8 {
			continue
		}
		if bytes.Equal(decodedData[:8], buyIdentifier) {
			// This is a Buy transaction.
			buyTransaction = &AMMBuyTransaction{
				BaseAmountOut:    binary.LittleEndian.Uint64(decodedData[8:16]),
				MaxQuoteAmountIn: binary.LittleEndian.Uint64(decodedData[16:24]),
			}

			instAccounts, err := i.ResolveInstructionAccounts(&tx.Message)
			if err != nil {
				slog.Error("Error resolving instruction accounts: " + err.Error())
				return nil, err
			}

			// fill the rest with the fields
			buyTransaction.Pool = instAccounts[poolIndex].PublicKey
			buyTransaction.User = instAccounts[userIndex].PublicKey
			buyTransaction.GlobalConfig = instAccounts[globalConfigIndex].PublicKey
			buyTransaction.BaseMint = instAccounts[baseMintIndex].PublicKey
			buyTransaction.QuoteMint = instAccounts[quoteMintIndex].PublicKey
			buyTransaction.UserBaseTokenAccount = instAccounts[userBaseTokenIndex].PublicKey
			buyTransaction.UserQuoteTokenAccount = instAccounts[userQuoteTokenIndex].PublicKey
			buyTransaction.PoolBaseTokenAccount = instAccounts[poolBaseTokenIndex].PublicKey
			buyTransaction.PoolQuoteTokenAccount = instAccounts[poolQuoteTokenIndex].PublicKey
			buyTransaction.ProtocolFeeRecipient = instAccounts[protocolFeeRecipientIndex].PublicKey
			buyTransaction.ProtocolFeeRecipientTokenAccount = instAccounts[protocolFeeRecipientTokenAccountIndex].PublicKey
			buyTransaction.CoinCreatorVaultATA = instAccounts[coinCreatorVaultATAIndex].PublicKey
			buyTransaction.CoinCreatorVaultAuthority = instAccounts[coinCreatorVaultAuthorityIndex].PublicKey

			return &AMMSwapTransaction{
				BuyTransaction: buyTransaction,
			}, nil
		} else if bytes.Equal(decodedData[:8], sellIdentifier) {
			// This is a Sell transaction.
			sellTransaction = &AMMSellTransaction{
				BaseAmountIn:      binary.LittleEndian.Uint64(decodedData[8:16]),
				MinQuoteAmountOut: binary.LittleEndian.Uint64(decodedData[16:24]),
			}

			instAccounts, err := i.ResolveInstructionAccounts(&tx.Message)
			if err != nil {
				slog.Error("Error resolving instruction accounts: " + err.Error())
				return nil, err
			}

			// fill the rest of the fields
			sellTransaction.Pool = instAccounts[poolIndex].PublicKey
			sellTransaction.User = instAccounts[userIndex].PublicKey
			sellTransaction.GlobalConfig = instAccounts[globalConfigIndex].PublicKey
			sellTransaction.BaseMint = instAccounts[baseMintIndex].PublicKey
			sellTransaction.QuoteMint = instAccounts[quoteMintIndex].PublicKey
			sellTransaction.UserBaseTokenAccount = instAccounts[userBaseTokenIndex].PublicKey
			sellTransaction.UserQuoteTokenAccount = instAccounts[userQuoteTokenIndex].PublicKey
			sellTransaction.PoolBaseTokenAccount = instAccounts[poolBaseTokenIndex].PublicKey
			sellTransaction.PoolQuoteTokenAccount = instAccounts[poolQuoteTokenIndex].PublicKey
			sellTransaction.ProtocolFeeRecipient = instAccounts[protocolFeeRecipientIndex].PublicKey
			sellTransaction.ProtocolFeeRecipientTokenAccount = instAccounts[protocolFeeRecipientTokenAccountIndex].PublicKey
			sellTransaction.CoinCreatorVaultATA = instAccounts[coinCreatorVaultATAIndex].PublicKey
			sellTransaction.CoinCreatorVaultAuthority = instAccounts[coinCreatorVaultAuthorityIndex].PublicKey

			return &AMMSwapTransaction{
				SellTransaction: sellTransaction,
			}, nil
		}
	}

	// Look for the Buy or Sell transaction in the inner instructions.
	for _, ii := range *innerInstructions {
		for _, i := range ii.Instructions {
			decodedData, err := base58.Decode(i.Data.String())
			if err != nil {
				continue
			}
			if len(decodedData) < 8 {
				continue
			}
			if bytes.Equal(decodedData[:8], buyIdentifier) {
				// This is a Buy transaction.
				buyTransaction = &AMMBuyTransaction{
					BaseAmountOut:    binary.LittleEndian.Uint64(decodedData[8:16]),
					MaxQuoteAmountIn: binary.LittleEndian.Uint64(decodedData[16:24]),
				}

				instAccounts, err := i.ResolveInstructionAccounts(&tx.Message)
				if err != nil {
					slog.Error("Error resolving instruction accounts: " + err.Error())
					return nil, err
				}

				// fill the rest of the fields
				buyTransaction.Pool = instAccounts[poolIndex].PublicKey
				buyTransaction.User = instAccounts[userIndex].PublicKey
				buyTransaction.GlobalConfig = instAccounts[globalConfigIndex].PublicKey
				buyTransaction.BaseMint = instAccounts[baseMintIndex].PublicKey
				buyTransaction.QuoteMint = instAccounts[quoteMintIndex].PublicKey
				buyTransaction.UserBaseTokenAccount = instAccounts[userBaseTokenIndex].PublicKey
				buyTransaction.UserQuoteTokenAccount = instAccounts[userQuoteTokenIndex].PublicKey
				buyTransaction.PoolBaseTokenAccount = instAccounts[poolBaseTokenIndex].PublicKey
				buyTransaction.PoolQuoteTokenAccount = instAccounts[poolQuoteTokenIndex].PublicKey
				buyTransaction.ProtocolFeeRecipient = instAccounts[protocolFeeRecipientIndex].PublicKey
				buyTransaction.ProtocolFeeRecipientTokenAccount = instAccounts[protocolFeeRecipientTokenAccountIndex].PublicKey
				buyTransaction.CoinCreatorVaultATA = instAccounts[coinCreatorVaultATAIndex].PublicKey
				buyTransaction.CoinCreatorVaultAuthority = instAccounts[coinCreatorVaultAuthorityIndex].PublicKey

				return &AMMSwapTransaction{
					BuyTransaction: buyTransaction,
				}, nil
			} else if bytes.Equal(decodedData[:8], sellIdentifier) {
				// This is a Sell transaction.
				sellTransaction = &AMMSellTransaction{
					BaseAmountIn:      binary.LittleEndian.Uint64(decodedData[8:16]),
					MinQuoteAmountOut: binary.LittleEndian.Uint64(decodedData[16:24]),
				}

				instAccounts, err := i.ResolveInstructionAccounts(&tx.Message)
				if err != nil {
					slog.Error("Error resolving instruction accounts: " + err.Error())
					return nil, err
				}

				// fill the rest of the fields
				sellTransaction.Pool = instAccounts[poolIndex].PublicKey
				sellTransaction.User = instAccounts[userIndex].PublicKey
				sellTransaction.GlobalConfig = instAccounts[globalConfigIndex].PublicKey
				sellTransaction.BaseMint = instAccounts[baseMintIndex].PublicKey
				sellTransaction.QuoteMint = instAccounts[quoteMintIndex].PublicKey
				sellTransaction.UserBaseTokenAccount = instAccounts[userBaseTokenIndex].PublicKey
				sellTransaction.UserQuoteTokenAccount = instAccounts[userQuoteTokenIndex].PublicKey
				sellTransaction.PoolBaseTokenAccount = instAccounts[poolBaseTokenIndex].PublicKey
				sellTransaction.PoolQuoteTokenAccount = instAccounts[poolQuoteTokenIndex].PublicKey
				sellTransaction.ProtocolFeeRecipient = instAccounts[protocolFeeRecipientIndex].PublicKey
				sellTransaction.ProtocolFeeRecipientTokenAccount = instAccounts[protocolFeeRecipientTokenAccountIndex].PublicKey
				sellTransaction.CoinCreatorVaultATA = instAccounts[coinCreatorVaultATAIndex].PublicKey
				sellTransaction.CoinCreatorVaultAuthority = instAccounts[coinCreatorVaultAuthorityIndex].PublicKey

				return &AMMSwapTransaction{
					SellTransaction: sellTransaction,
				}, nil
			}
		}
	}

	return nil, errors.New("no AMM transaction found")
}
