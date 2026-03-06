package pumpfunsdk

import (
	"context"
	"errors"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gookit/slog"
	"github.com/near/borsh-go"
	"time"
)

// This file is called api, since it uses requests to gather more data.

const (
	retries = 10  // number of retries
	delay   = 100 // ms
)

// ClientOpts is a struct that holds the client options.
type ClientOpts struct {
	Retries    int   // number of Retries
	Delay      int64 // ms
	Commitment rpc.CommitmentType
}

// GetBondingCurve returns the bonding curve data for the given bonding curve key.
// The bonding curve key is the public key of the account that holds the bonding curve data.
// The bonding curve data is deserialized into a *BondingCurve.
// This uses the process commitment level.
// If the client is nil, or the bondingCurveKey is zero, an error is returned.
//
// opts : ClientOpts - optional client options.
func GetBondingCurve(client Client, bondingCurveKey solana.PublicKey, opts ...ClientOpts) (*BondingCurve, error) {
	if client == nil || bondingCurveKey.IsZero() {
		return nil, errors.New("client or bonding_curve_key is nil")
	}

	// Check if the client options are provided.
	var clientOpts ClientOpts
	if len(opts) > 0 {
		clientOpts = opts[0]
	} else {
		clientOpts = ClientOpts{
			Retries:    retries,
			Delay:      delay,
			Commitment: rpc.CommitmentConfirmed,
		}
	}

	var out *rpc.GetAccountInfoResult
	var err error

	for i := 0; i < clientOpts.Retries; i++ {
		out, err = client.GetAccountInfoWithOpts(context.TODO(),
			bondingCurveKey,
			&rpc.GetAccountInfoOpts{
				Commitment: clientOpts.Commitment,
			},
		)

		if err == nil && out.Value != nil && out.Value.Data != nil && len(out.Value.Data.GetBinary()) > 0 {
			break
		}

		time.Sleep(time.Duration(clientOpts.Delay) * time.Millisecond)
	}

	if err != nil {
		return nil, errors.New("Retries exhausted for address: " + bondingCurveKey.String())
	}
	if out == nil {
		return nil, errors.New("account info not found")
	}
	if out.Value == nil {
		return nil, errors.New("account not found")
	}
	if out.Value.Data == nil {
		return nil, errors.New("account data not found")
	}

	data := out.Value.Data.GetBinary()
	if len(data) == 0 {
		return nil, errors.New("account data is empty")
	}

	y := new(BondingCurve)
	err = borsh.Deserialize(y, data[8:])
	if err != nil {
		slog.Error("Error deserializing data: " + err.Error())
		return nil, err
	}
	return y, nil
}
