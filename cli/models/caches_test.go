package models

import (
	"Zed/mocks"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestFirstTimeBuyCache_Add(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		args    args
		wantErr bool
	}{
		{args{"test"}, false},
		{args{""}, true},
	}
	for _, tt := range tests {
		c := NewFirstTimeBuyCache()
		if err := c.Add(tt.args.key); (err != nil) != tt.wantErr {
			t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
		}

	}
	t.Run("Add multiple times", func(t *testing.T) {
		c := NewFirstTimeBuyCache()
		if err := c.Add("test"); err != nil {
			t.Fatal(err)
		}
		if err := c.Add("test"); err != nil {
			t.Fatal(err)
		}
		if err := c.Add(""); err == nil {
			t.Fatal(err)
		}
		if c.Size() != 1 {
			t.Fatal("size is not 1")
		}
	})
}

func TestFirstTimeBuyCache_Exists(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		args    args
		want    bool
		wantErr bool
	}{
		{args{"test"}, false, false},
		{args{""}, false, false},
	}
	for _, tt := range tests {
		c := NewFirstTimeBuyCache()
		if got := c.Exists(tt.args.key); got != tt.want {
			t.Errorf("Exists() = %v, want %v", got, tt.want)
		}
	}
	t.Run("Add and check", func(t *testing.T) {
		c := NewFirstTimeBuyCache()
		if ex := c.Exists("A"); ex {
			t.Fatal("found key that shouldn't be in cache")
		}
		if err := c.Add("A"); err != nil {
			t.Fatal(err)
		}
		if ex := c.Exists("A"); !ex {
			t.Fatal("didnt find key that should be in cache")
		}
		if ex := c.Exists("B"); ex {
			t.Fatal("found key that shouldn't be in cache")
		}
	})
}

func TestFirstTimeBuyCache_Preload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockIClient(ctrl)

	// Define the mock expectation for GetTokenAccountsByOwner
	mockClient.EXPECT().GetTokenAccountsByOwner(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&rpc.GetTokenAccountsResult{Value: make([]*rpc.TokenAccount, 0)}, nil).AnyTimes()

	type args struct {
		client IClient
		wallet string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"nil client", args{nil, ""}, true},
		{"nil wallet", args{mockClient, ""}, true},
		{"empty accounts", args{mockClient, "yxxiSEmGhm8mTjbeUZ8BgqoWz6oyMRG1WNnvWJSQJyZ"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewFirstTimeBuyCache()
			if err := c.Preload(tt.args.client, tt.args.wallet); (err != nil) != tt.wantErr {
				t.Errorf("Preload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	t.Run("Preload", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockClient := mocks.NewMockIClient(ctrl)
		c := NewFirstTimeBuyCache()
		// Mock call : GetTokenAccountsByOwner, it should return a list of accounts
		mockClient.EXPECT().GetTokenAccountsByOwner(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&rpc.GetTokenAccountsResult{
			Value: []*rpc.TokenAccount{
				{Pubkey: solana.MustPublicKeyFromBase58("9ssS4bUdrJm6wjjBvVejxXU9FWYCuHfovFo4EUsU7QGt")},
				{Pubkey: solana.MustPublicKeyFromBase58("8ssS4bUdrJm6wjjBvVejxXU9FWYCuHfovFo4EUsU7QGt")},
				{Pubkey: solana.MustPublicKeyFromBase58("7ssS4bUdrJm6wjjBvVejxXU9FWYCuHfovFo4EUsU7QGt")},
				{Pubkey: solana.MustPublicKeyFromBase58("6ssS4bUdrJm6wjjBvVejxXU9FWYCuHfovFo4EUsU7QGt")},
			},
		}, nil)
		if err := c.Preload(mockClient, "ZDK6SbnjEtgTWG1CqHhNueEFvziCcMQYQhW2UwJ3EzU"); err != nil {
			t.Fatal(err)
		}
		if c.Size() != 4 {
			t.Fatal("didnt get all ATAs from wallet")
		}
		if !c.Exists("9ssS4bUdrJm6wjjBvVejxXU9FWYCuHfovFo4EUsU7QGt") {
			t.Fatal("didnt find ATA")
		}
		if !c.Exists("8ssS4bUdrJm6wjjBvVejxXU9FWYCuHfovFo4EUsU7QGt") {
			t.Fatal("didnt find ATA")
		}
		if !c.Exists("7ssS4bUdrJm6wjjBvVejxXU9FWYCuHfovFo4EUsU7QGt") {
			t.Fatal("didnt find ATA")
		}
		if !c.Exists("6ssS4bUdrJm6wjjBvVejxXU9FWYCuHfovFo4EUsU7QGt") {
			t.Fatal("didnt find ATA")
		}
	})
}

// tokenAmount, exists := tokensOwnedCache.Get(mint)
func TestTokensOwnedCache_Preload(t *testing.T) {
	//wallet := "ZDK6SbnjEtgTWG1CqHhNueEFvziCcMQYQhW2UwJ3EzU"
	//cache := NewTokensOwnedCache()
	//err := cache.Preload(mocks.RealTestingClient, wallet)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//tokenAmount, exists := cache.Get("2J8L9hdrGE8Lt5EURhg62gNLymGf7onibFJX2MVxpump")
	//if !exists {
	//	t.Fatal("token not found")
	//}
	//if tokenAmount == 0 {
	//	t.Fatal("token amount is 0")
	//}
}
