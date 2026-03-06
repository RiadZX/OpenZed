package pumpfuncopytrader

import (
	"Zed/models"
	"Zed/pumpfunsdk"
	"errors"
)

type PurchaseSellData struct {
	pumpfunsdk.TradeEventData
	BondingCurve           string
	AssociatedBondingCurve string
	PreTokenBalance        float64 //in tokens already , so in case of pumpfun multiplied by 1e6 already
	PostTokenBalance       float64 //in tokens already , so in case of pumpfun multiplied by 1e6 already
}

// isFirstTimeBuyingCache checks if the ATA is in the cache
func isFirstTimeBuyingCache(associatedUserTokenAccount string, ftbCache *models.FirstTimeBuyCache) (bool, error) {
	if associatedUserTokenAccount == "" {
		return false, errors.New("ata is nil")
	}
	if ftbCache == nil {
		return false, errors.New("cache is nil")
	}
	existsInCache := ftbCache.Exists(associatedUserTokenAccount)
	if existsInCache {
		//Cache hit
		return false, nil
	}
	//Cache miss
	return true, nil
}
