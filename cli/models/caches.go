package models

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gookit/slog"
	"strconv"
	"sync"
	"time"
)

var (
	TOKEN_PROGRAM = solana.MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
)

func RefreshCachesWithWallet(client IClient, wallet string, config *CopyTraderConfig) {
	for _, task := range config.ActiveTasks {
		//preload first time buy cache
		err := task.FirstTimeBuyCache.Preload(client, wallet)
		if err != nil {
			slog.Error("Failed preloading first time buy cache")
			slog.Debug(spew.Sdump(err))
		}
		//preload tokens owned cache
		err = task.TokensOwnedCache.Preload(client, wallet)
		if err != nil {
			slog.Error("Failed preloading tokens owned cache")
			slog.Debug(spew.Sdump(err))
		}
	}
}

// FirstTimeBuyCache for storing if it's the first time an ATA is buying.
// Using a hash set approach where the key is the ATA and the value is an empty struct.
type FirstTimeBuyCache struct {
	cache map[string]struct{}
	lock  sync.RWMutex
	size  int
}

// NewFirstTimeBuyCache creates a new cache instance.
func NewFirstTimeBuyCache() FirstTimeBuyCache {
	return FirstTimeBuyCache{
		cache: make(map[string]struct{}),
	}
}

func (c *FirstTimeBuyCache) Size() int {
	if c == nil || c.cache == nil {
		return 0
	}
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.size
}

// Add inserts an ATA into the cache.
func (c *FirstTimeBuyCache) Add(key string) error {
	if c == nil || c.cache == nil || key == "" {
		return errors.New("cache is nil")
	}
	if c.Exists(key) {
		return nil
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	c.cache[key] = struct{}{}
	c.size++
	return nil
}

// Exists checks if an ATA is in the cache.
func (c *FirstTimeBuyCache) Exists(key string) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	_, found := c.cache[key]
	return found
}

func (c *FirstTimeBuyCache) Preload(client IClient, wallet string) error {
	var retries = 20
	if client == nil {
		return errors.New("client is nil")
	}
	if wallet == "" {
		return errors.New("wallet is empty")
	}
	//get all token accounts owned by the wallet
	//for each token account add it to the cache
	for i := 0; i < retries; i++ {
		accs, err := client.GetTokenAccountsByOwner(
			context.Background(),
			solana.MustPublicKeyFromBase58(wallet),
			&rpc.GetTokenAccountsConfig{
				Mint:      nil,
				ProgramId: &TOKEN_PROGRAM,
			},
			&rpc.GetTokenAccountsOpts{
				Commitment: "processed",
				Encoding:   "jsonParsed",
				DataSlice:  nil,
			},
		)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		//clear cache
		c.cache = make(map[string]struct{})
		c.size = 0

		for _, acc := range accs.Value {
			ata := acc.Pubkey
			err := c.Add(ata.String())
			if err != nil {
				return err
			}
		}
		slog.Debug("Preloaded cache with" + strconv.Itoa(len(c.cache)) + "entries")
		break
	}
	return nil
}

/* -----Tokens account cache------ */

type GeneratedAccountStruct struct {
	Parsed struct {
		Info struct {
			IsNative    bool   `json:"isNative"`
			Mint        string `json:"mint"`
			Owner       string `json:"owner"`
			State       string `json:"state"`
			TokenAmount struct {
				Amount         string  `json:"amount"`
				Decimals       int     `json:"decimals"`
				UIAmount       float64 `json:"uiAmount"`
				UIAmountString string  `json:"uiAmountString"`
			} `json:"tokenAmount"`
		} `json:"info"`
		Type string `json:"type"`
	} `json:"parsed"`
	Program string `json:"program"`
	Space   int    `json:"space"`
}

type TokensOwnedCache struct {
	cache map[string]float64
	lock  sync.RWMutex
}

func NewTokensOwnedCache() TokensOwnedCache {
	return TokensOwnedCache{
		cache: make(map[string]float64),
	}
}

func (c *TokensOwnedCache) Add(key string, amount float64) {
	if c == nil || c.cache == nil || key == "" {
		return
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	c.cache[key] = amount
}

func (c *TokensOwnedCache) Get(key string) (tokenamount float64, exists bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	amount, found := c.cache[key]
	return amount, found
}

func (c *TokensOwnedCache) Exists(key string) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	_, found := c.cache[key]
	return found
}

func (c *TokensOwnedCache) Remove(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.cache, key)
}

func (c *TokensOwnedCache) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.cache = make(map[string]float64)
}

func (c *TokensOwnedCache) Preload(client IClient, wallet string) error {
	//get all tokens owned by wallet + add the amount to the cache
	var retries = 20
	var err error
	if client == nil {
		return errors.New("client is nil")
	}
	if wallet == "" {
		return errors.New("wallet is empty")
	}
	//get all token accounts owned by the wallet
	//for each token account add it to the cache
	for i := 0; i < retries; i++ {
		accs, err := client.GetTokenAccountsByOwner(
			context.Background(),
			solana.MustPublicKeyFromBase58(wallet),
			&rpc.GetTokenAccountsConfig{
				Mint:      nil,
				ProgramId: &TOKEN_PROGRAM,
			},
			&rpc.GetTokenAccountsOpts{
				Commitment: "processed",
				Encoding:   "jsonParsed",
				DataSlice:  nil,
			},
		)
		if err != nil {
			slog.Debug("Error reloading tokens owned cache: ", err)
			time.Sleep(2 * time.Second)
			continue
		}

		//clear cache
		c.Clear()

		for _, acc := range accs.Value {
			jsonbytes, err := acc.Account.Data.MarshalJSON()
			if err != nil {
				spew.Dump(acc)
				return err
			}
			//json bytes to struct with json tags
			var accStruct GeneratedAccountStruct
			err = json.Unmarshal(jsonbytes, &accStruct)
			if err != nil {
				return err
			}
			if accStruct.Parsed.Info.TokenAmount.UIAmount == 0 {
				continue
			}

			c.Add(accStruct.Parsed.Info.Mint, accStruct.Parsed.Info.TokenAmount.UIAmount)
		}
		slog.Debug("Preloaded cache with " + strconv.Itoa(len(c.cache)) + " entries")
		break
	}
	if err != nil {
		slog.Error("Failed preloading cache")
		slog.Debug(spew.Sdump(err))
		return err
	}
	return nil
}
