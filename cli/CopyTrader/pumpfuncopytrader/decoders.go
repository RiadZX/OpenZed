package pumpfuncopytrader

import (
	"Zed/pumpfunsdk"
	"bytes"
	"errors"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/mr-tron/base58/base58"
)

func getPurchaseData(t *pumpfunsdk.TradeEventData, tx *solana.Transaction, innerInstructions *[]rpc.InnerInstruction) (PurchaseSellData, error) {
	if t == nil {
		return PurchaseSellData{}, errors.New("TradeEventData is nil")
	}
	pd := PurchaseSellData{
		TradeEventData:         *t,
		BondingCurve:           "",
		AssociatedBondingCurve: "",
	}

	var targetSequence []byte
	if t.IsBuy {
		targetSequence = []byte{102, 6, 61, 18, 1, 218}
	} else {
		targetSequence = []byte{0x33, 0xe6, 0x85, 0xa4, 0x01, 0x7f}
	}

	for _, i := range tx.Message.Instructions {
		decodedData, err := base58.Decode(i.Data.String())
		if err != nil {
			continue
		}
		is := i.Data.String()
		if len(is) >= 8 && bytes.Equal(decodedData[:6], targetSequence) {
			indexOfBondingCurve := i.Accounts[3]
			if len(tx.Message.AccountKeys) <= int(indexOfBondingCurve) {
				continue
			}
			pd.BondingCurve = tx.Message.AccountKeys[indexOfBondingCurve].String()

			indexOfAssociatedBondingCurve := i.Accounts[4]
			if len(tx.Message.AccountKeys) <= int(indexOfAssociatedBondingCurve) {
				continue
			}
			pd.AssociatedBondingCurve = tx.Message.AccountKeys[indexOfAssociatedBondingCurve].String()
		}
	}

	if pd.BondingCurve == "" || pd.AssociatedBondingCurve == "" {
		//Search through the inner instructions
		for _, i := range *innerInstructions {
			for _, ii := range i.Instructions {
				decodedData, err := base58.Decode(ii.Data.String())
				if err != nil {
					continue
				}
				is := ii.Data.String()
				if len(is) >= 8 && bytes.Equal(decodedData[:6], targetSequence) {
					indexOfBondingCurve := ii.Accounts[3]
					if len(tx.Message.AccountKeys) <= int(indexOfBondingCurve) {
						continue
					}
					pd.BondingCurve = tx.Message.AccountKeys[indexOfBondingCurve].String()

					indexOfAssociatedBondingCurve := ii.Accounts[4]
					if len(tx.Message.AccountKeys) <= int(indexOfAssociatedBondingCurve) {
						continue
					}
					pd.AssociatedBondingCurve = tx.Message.AccountKeys[indexOfAssociatedBondingCurve].String()
					break
				}
			}
		}
	}

	//if still not found then error
	if pd.BondingCurve == "" || pd.AssociatedBondingCurve == "" {
		return pd, errors.New("bonding curve or Associated Bonding curve not found")
	}

	return pd, nil
}
