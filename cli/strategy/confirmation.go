package strategy

import (
	"sync"
)

/*
This file will handle all things related to confirming transactions.
Holding the buffer of sent transactions and interacting with them.
The success/fail manager of it.

Data structure for saving a map of sent signatures. Thread safe.
Enum to define the status of a transaction.
Function to parse a transaction, and decide if its a success or fail.
*/

// TransactionStatus is the status of a transaction.
type TransactionStatus int

var (
	// TransactionError is the status of a transaction that failed.
	TransactionError TransactionStatus = -1

	// TransactionPending is the status of a transaction that is still pending, and thus the default status.
	TransactionPending TransactionStatus = 0

	// TransactionSuccess is the status of a transaction that succeeded.
	TransactionSuccess TransactionStatus = 1

	// SigBuffer Global buffer to hold the sent transactions.
	SigBuffer *SignaturesBuffer
)

func init() {
	// Initialize the buffer.
	SigBuffer = NewSignaturesBuffer()
}

// SignaturesBuffer is a thread safe buffer to hold the sent transactions.
type SignaturesBuffer struct {
	// Signatures is a map of the sent transactions.
	Signatures map[string]TransactionStatus

	// Mutex is the lock for the map.
	Mutex *sync.Mutex
}

// NewSignaturesBuffer creates a new SignaturesBuffer.
func NewSignaturesBuffer() *SignaturesBuffer {
	return &SignaturesBuffer{
		Signatures: make(map[string]TransactionStatus),
		Mutex:      &sync.Mutex{},
	}
}

// Append adds a signature to the buffer.
func (sb *SignaturesBuffer) Append(signature string) {
	sb.Mutex.Lock()
	defer sb.Mutex.Unlock()
	sb.Signatures[signature] = TransactionPending
}

// SetStatusAndLog sets the status of a signature and logs it.
func (sb *SignaturesBuffer) SetStatusAndLog(signature string, status TransactionStatus) {
	sb.Mutex.Lock()
	defer sb.Mutex.Unlock()
	//if it was not set the nignore
	if _, ok := sb.Signatures[signature]; !ok {
		return
	}
	sb.Signatures[signature] = status
}

// GetStatus gets the status of a signature.
func (sb *SignaturesBuffer) GetStatus(signature string) TransactionStatus {
	sb.Mutex.Lock()
	defer sb.Mutex.Unlock()
	// If the signature is not found, return pending.
	if status, ok := sb.Signatures[signature]; ok {
		return status
	} else {
		return TransactionPending
	}
}
