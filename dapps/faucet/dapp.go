package faucet

import (
	faucet "github.com/iotaledger/goshimmer/dapps/faucet/packages"
	"github.com/iotaledger/goshimmer/packages/binary/messagelayer/message"
	"github.com/iotaledger/goshimmer/packages/binary/messagelayer/tangle"
	"github.com/iotaledger/goshimmer/plugins/messagelayer"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
)

const name = "Faucet" // name of the plugin

var App = node.NewPlugin(name, node.Disabled, configure, run)

var log *logger.Logger

func configure(*node.Plugin) {
	log = logger.NewLogger(name)
	faucet.ConfigureFaucet()
	configureEvents()
}

func configureEvents() {
	messagelayer.Tangle.Events.MessageSolid.Attach(events.NewClosure(func(cachedTransaction *message.CachedMessage, cachedTransactionMetadata *tangle.CachedMessageMetadata) {
		defer cachedTransaction.Release()
		cachedTransactionMetadata.Release()

		if msg := cachedTransaction.Unwrap(); msg != nil {
			if faucet.IsFaucetReq(msg) {
				faucet.SendFunds(msg)
			}
		} else {
			log.Errorf("Fail to unwrap cachedTransaction")
		}
	}))
}

func run(*node.Plugin) {}