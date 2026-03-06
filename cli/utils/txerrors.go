package utils

import (
	"github.com/gagliardetto/solana-go/rpc"
	"strings"
)

// This file contains the most common errors and how to handle them.
// Some Errors Stop the Sending, while others should be logged and ignored.
// This file will be used in all places where transactions be sent, to handle errors.

// Function that will handle the simulated transaction errors.
// Some errors are critical and should stop the sending of the transaction.
// Others are not critical and should be logged and ignored.

// Create types of errors to return, criticalERror, ignoreable etc.
type TxErrorType int

const (
	// CriticalError should stop the sending of the transaction.
	CriticalError TxErrorType = iota
	// IgnoreableError should be logged and ignored.
	IgnoreableError
	// NoError should be returned if no error is found.
	NoError

	// prefix for the error messages. cErr = Critical Error, iErr = Ignoreable Error.
	// Define common errors, the function will return the type, and if known the error message.

	CErrInsufficientFunds  = "Insufficient funds"                                   // Critical
	IErrIncorrectProgramId = "Incorrect program"                                    // Ignoreable, since this only happens on simulation.
	CErrSlippage           = "Slippage"                                             // Critical, this means that it failed due to slippage.
	IAMMError              = "The amm account owner is not match with this program" // Ignorable, this means that the AMM failed.
	CErrUnknown            = "Unknown error"                                        // Critical, this means that the error is unknown.
)

// GetSimulationErrorType returns the type of error and the error message.
func GetSimulationErrorType(simulationResult *rpc.SimulateTransactionResult) (TxErrorType, string) {
	if simulationResult == nil {
		return CriticalError, "Simulation result is nil"
	}
	// If there is no error, return NoError.
	if simulationResult.Err == nil {
		return NoError, ""
	}
	// Some errors require looking through the logs.
	if simulationResult.Logs != nil {
		for _, log := range simulationResult.Logs {
			// 0x1 is the error code for insufficient funds.
			if strings.Contains(log, "Error: The amm account owner is not match with this program") {
				return IgnoreableError, IAMMError
			}
			if strings.Contains(log, "custom program error: 0x1") && !strings.Contains(log, "custom program error: 0x1b") {
				return CriticalError, CErrInsufficientFunds
			}
			// Program log: Error: IncorrectProgramId -> This is a common error when the program id is incorrect.
			if strings.Contains(log, "Error: IncorrectProgramId") {
				return IgnoreableError, IErrIncorrectProgramId // This is only a simulation error, only happens during simulation.
			}
			// PumpFun, slippage error.
			if strings.Contains(log, "Error Message: slippage") {
				return CriticalError, CErrSlippage
			}
		}
	} else {
		return CriticalError, "Simulation result logs are nil"
	}
	// This is an unknown error
	return CriticalError, CErrUnknown
}

// IsIgnorableError returns true if the error is ignorable, so we can ignore it and continue with the transaction.
func IsIgnorableError(err error) bool {
	if err == nil {
		return true
	}
	if strings.Contains(err.Error(), "nil response") {
		return true
	}
	if strings.Contains(err.Error(), CErrInsufficientFunds) {
		return false
	}
	if strings.Contains(err.Error(), IErrIncorrectProgramId) {
		return true
	}
	if strings.Contains(err.Error(), CErrSlippage) {
		return false
	}
	if strings.Contains(err.Error(), IAMMError) {
		return true
	}
	if strings.Contains(err.Error(), CErrUnknown) {
		return true
	}
	return false
}
