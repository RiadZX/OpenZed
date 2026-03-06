package models

/* This file contains the structs that will be used to:
- Log data to discord webhooks
- Log data to the database

This is generic, and can be used for any solana program.
*/

// SwapInfo is the struct that will be used to log swap data to the database.
// The same struct is also used for discord webhooks.
type SwapInfo struct {
	ID                   string  `json:"id,omitempty"` // Cosmos DB requires an ID field
	License              string  `json:"LicenseKey"`
	Timestamp            string  `json:"Timestamp"`
	UserWallet           string  `json:"UserWallet"`
	ProgramName          string  `json:"ProgramName"` // Currently: PumpFun / Raydium
	IsBuy                bool    `json:"IsBuy"`
	TokenMint            string  `json:"TokenMint"`
	TokenMintAmount      float64 `json:"TokenMintAmount"`
	SolanaAmount         float64 `json:"SolanaAmount"`
	TransactionSignature string  `json:"TransactionSignature"`
}
