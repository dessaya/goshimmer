package transactionspammer

import (
	"sync"
	"time"

	"github.com/iotaledger/goshimmer/packages/daemon"
	"github.com/iotaledger/goshimmer/packages/model/value_transaction"
	"github.com/iotaledger/goshimmer/plugins/gossip"
	"github.com/iotaledger/goshimmer/plugins/tipselection"
)

var spamming = false

var startMutex sync.Mutex

var shutdownSignal chan int

var sentCounter = uint(0)

func Start(tps uint) {
	startMutex.Lock()

	if !spamming {
		shutdownSignal = make(chan int, 1)

		func(shutdownSignal chan int) {
			daemon.BackgroundWorker("Transaction Spammer", func() {
				for {
					start := time.Now()
					totalSentCounter := int64(0)

					for {
						select {
						case <-daemon.ShutdownSignal:
							return

						case <-shutdownSignal:
							return

						default:
							sentCounter++
							totalSentCounter++

							tx := value_transaction.New()
							tx.SetValue(totalSentCounter)
							tx.SetBranchTransactionHash(tipselection.GetRandomTip())
							tx.SetTrunkTransactionHash(tipselection.GetRandomTip())

							gossip.Events.ReceiveTransaction.Trigger(tx.MetaTransaction)

							if sentCounter >= tps {
								duration := time.Since(start)
								if duration < time.Second {
									time.Sleep(time.Second - duration)

									start = time.Now()
								}

								sentCounter = 0
							}
						}
					}
				}
			})
		}(shutdownSignal)

		spamming = true
	}

	startMutex.Unlock()
}

func Stop() {
	startMutex.Lock()

	if spamming {
		close(shutdownSignal)

		spamming = false
	}

	startMutex.Unlock()
}

func GenerateBundle(bundleLength int) (result []*value_transaction.ValueTransaction) {
	result = make([]*value_transaction.ValueTransaction, bundleLength)

	branch := tipselection.GetRandomTip()
	trunk := tipselection.GetRandomTip()

	for i := 0; i < bundleLength; i++ {
		tx := value_transaction.New()
		tx.SetTail(i == 0)
		tx.SetHead(i == bundleLength - 1)
		tx.SetTimestamp(sentCounter)
		tx.SetBranchTransactionHash(branch)
		tx.SetTrunkTransactionHash(trunk)

		result[i] = tx

		trunk = tx.GetHash()
	}

	return result
}
