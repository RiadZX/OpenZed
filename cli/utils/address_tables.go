package utils

import (
	"Zed/models"
	"context"
	"errors"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gookit/slog"
	"sync"
	"time"

	"github.com/gagliardetto/solana-go"
	addresslookuptable "github.com/gagliardetto/solana-go/programs/address-lookup-table"
)

var (
	accountInfoCache = make(map[string]rpc.GetAccountInfoResult)
	accountInfoMutex = sync.RWMutex{}
)

func ResolveAddressTables(client models.IClient, tx *solana.Transaction) (*solana.Transaction, error) {
	if !tx.Message.IsVersioned() {
		return tx, nil
	}

	tblKeys := tx.Message.GetAddressTableLookups().GetTableIDs()
	if len(tblKeys) == 0 {
		return tx, nil
	}

	numLookups := tx.Message.GetAddressTableLookups().NumLookups()
	if numLookups == 0 {
		return tx, nil
	}

	resolutions := make(map[solana.PublicKey]solana.PublicKeySlice)
	for _, key := range tblKeys {
		accountInfoMutex.RLock()
		if _, ok := accountInfoCache[key.String()]; !ok {
			accountInfoMutex.RUnlock()
			accountInfoMutex.Lock()
			info, err := getAccountInfo(client, key.String(), 10, 50)
			if err != nil || info == nil {
				accountInfoMutex.Unlock()
				return nil, err
			}
			accountInfoCache[key.String()] = *info
			accountInfoMutex.Unlock()
			accountInfoMutex.RLock()
		}
		result := accountInfoCache[key.String()]
		accountInfoMutex.RUnlock()

		tableContent, err := addresslookuptable.DecodeAddressLookupTableState((&result).GetBinary())
		if err != nil {
			slog.Error("Error decoding address lookup table: " + err.Error())
			return nil, err
		}

		resolutions[key] = tableContent.Addresses
	}
	err := tx.Message.SetAddressTables(resolutions)
	if err != nil {
		return nil, err
	}

	err = tx.Message.ResolveLookups()
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func getAccountInfo(client models.IClient, address string, delay int64, retries int64) (*rpc.GetAccountInfoResult, error) {
	if retries == 0 {
		return nil, errors.New("retries exhausted for address: " + address)
	}
	if client == nil || address == "" {
		return nil, errors.New("client or address is nil")
	}

	out, err := client.GetAccountInfoWithOpts(context.TODO(),
		solana.MustPublicKeyFromBase58(address),
		&rpc.GetAccountInfoOpts{
			Commitment: rpc.CommitmentProcessed,
		},
	)
	if err != nil {
		time.Sleep(time.Duration(delay*(50-retries)) * time.Millisecond)
		return getAccountInfo(client, address, delay, retries-1)
	}
	if out.Value == nil {
		return nil, errors.New("account not found")
	}
	if out.Value.Data == nil {
		return nil, errors.New("account data not found")
	}
	return out, nil
}
