package utxodb

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
)

type pendingTransaction struct {
	confirmDeadline time.Time
	tx              *transaction.Transaction
	hasConflicts    bool
	onConfirm       func()
}

// implements valuetangle.ValueTangle by wrapping UTXODB and adding a fake confirmation delay
type ConfirmEmulator struct {
	UtxoDB                 *UtxoDB
	confirmTime            time.Duration
	randomize              bool
	confirmFirstInConflict bool
	pendingTransactions    map[transaction.ID]*pendingTransaction
	mutex                  sync.Mutex
}

func NewConfirmEmulator(confirmTime time.Duration, randomize bool, confirmFirstInConflict bool) *ConfirmEmulator {
	ce := &ConfirmEmulator{
		UtxoDB:                 New(),
		pendingTransactions:    make(map[transaction.ID]*pendingTransaction),
		confirmTime:            confirmTime,
		randomize:              randomize,
		confirmFirstInConflict: confirmFirstInConflict,
	}
	go ce.confirmLoop()
	return ce
}

func (ce *ConfirmEmulator) PostTransaction(tx *transaction.Transaction, onConfirm func()) error {
	if onConfirm == nil {
		onConfirm = func() {}
	}
	ce.mutex.Lock()
	defer ce.mutex.Unlock()

	if ce.confirmTime == 0 {
		if err := ce.UtxoDB.AddTransaction(tx); err != nil {
			return err
		}
		onConfirm()
		fmt.Printf("utxodb.ConfirmEmulator CONFIRMED IMMEDIATELY: %s\n", tx.ID().String())
		return nil
	}
	if err := ce.UtxoDB.ValidateTransaction(tx); err != nil {
		return err
	}
	for txid, ptx := range ce.pendingTransactions {
		if AreConflicting(tx, ptx.tx) {
			ptx.hasConflicts = true
			return fmt.Errorf("utxodb.ConfirmEmulator rejected: new tx %s conflicts with pending tx %s", tx.ID().String(), txid.String())
		}
	}
	var confTime time.Duration
	if ce.randomize {
		confTime = time.Duration(rand.Int31n(int32(ce.confirmTime)) + int32(ce.confirmTime)/2)
	} else {
		confTime = ce.confirmTime
	}
	deadline := time.Now().Add(confTime)

	ce.pendingTransactions[tx.ID()] = &pendingTransaction{
		confirmDeadline: deadline,
		tx:              tx,
		hasConflicts:    false,
		onConfirm:       onConfirm,
	}
	fmt.Printf("utxodb.ConfirmEmulator ADDED PENDING TRANSACTION: %s\n", tx.ID().String())
	return nil
}

const loopPeriod = 500 * time.Millisecond

func (ce *ConfirmEmulator) confirmLoop() {
	maturedTxs := make([]transaction.ID, 0)
	for {
		time.Sleep(loopPeriod)

		maturedTxs = maturedTxs[:0]
		nowis := time.Now()
		ce.mutex.Lock()

		for txid, ptx := range ce.pendingTransactions {
			if ptx.confirmDeadline.Before(nowis) {
				maturedTxs = append(maturedTxs, txid)
			}
		}

		if len(maturedTxs) == 0 {
			ce.mutex.Unlock()
			continue
		}

		for _, txid := range maturedTxs {
			ptx := ce.pendingTransactions[txid]
			if ptx.hasConflicts && !ce.confirmFirstInConflict {
				// do not confirm if tx has conflicts
				fmt.Printf("!!! utxodb.ConfirmEmulator: rejected because has conflicts %s\n", txid.String())
				continue
			}
			if err := ce.UtxoDB.AddTransaction(ptx.tx); err != nil {
				fmt.Printf("!!!! utxodb.AddTransaction: %v\n", err)
			} else {
				ptx.onConfirm()
				fmt.Printf("+++ utxodb.ConfirmEmulator: CONFIRMED %s after %v\n", txid.String(), ce.confirmTime)
			}
		}

		for _, txid := range maturedTxs {
			delete(ce.pendingTransactions, txid)
		}
		ce.mutex.Unlock()
	}
}

func (ce *ConfirmEmulator) GetAddressOutputs(addr address.Address) (map[transaction.OutputID][]*balance.Balance, error) {
	return ce.UtxoDB.GetAddressOutputs(addr), nil
}

func (ce *ConfirmEmulator) IsConfirmed(txid *transaction.ID) (bool, error) {
	return ce.UtxoDB.IsConfirmed(txid), nil
}

func (ce *ConfirmEmulator) GetTransaction(id transaction.ID) (*transaction.Transaction, error) {
	tx, _ := ce.UtxoDB.GetTransaction(id)
	return tx, nil
}

func (ce *ConfirmEmulator) RequestFunds(target address.Address) (*transaction.Transaction, error) {
	return ce.UtxoDB.RequestFunds(target)
}
