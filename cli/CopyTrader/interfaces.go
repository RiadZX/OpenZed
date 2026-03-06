package copytrader

import (
	"Zed/db"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc/ws"
	geyser_pb "github.com/weeaa/goyser/pb"
)

type CopyTrader interface {

	// Init initializes the copy trading bot.
	// Returns an error if initialization fails.
	Init() error

	// Start begins the copy trading processes.
	// Returns an error if the process fails to start.
	Start() error

	// HandleWSMessage processes a WebSocket message.
	// Parameters:
	// - got: The WebSocket log result.
	// - blockhash: The most recent block hash to use.
	HandleWSMessage(got *ws.LogResult, blockhash *solana.Hash)

	// HandleGRPCMessage processes a gRPC message.
	// Parameters:
	// - geyserTx: The gRPC transaction update.
	// - blockhash: The most recent block hash to use.
	HandleGRPCMessage(geyserTx *geyser_pb.SubscribeUpdateTransaction, blockhash *solana.Hash)

	// GetProgramName returns the name of the trading program.
	// Returns a string representing the program name.
	GetProgramName() string

	// SetDatabase sets the database connection for the copy trading bot.
	// Parameters:
	// - db: The database connection.
	SetDatabase(db *db.Connection)
}
