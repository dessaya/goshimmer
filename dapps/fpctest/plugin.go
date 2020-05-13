package fpctest

import (
	"math/rand"
	"time"

	fpcTestPayload "github.com/iotaledger/goshimmer/dapps/fpctest/packages/payload"
	"github.com/iotaledger/goshimmer/dapps/fpctest/packages/tangle"
	"github.com/iotaledger/goshimmer/packages/binary/messagelayer/message"
	messageTangle "github.com/iotaledger/goshimmer/packages/binary/messagelayer/tangle"
	"github.com/iotaledger/goshimmer/packages/database"
	"github.com/iotaledger/goshimmer/packages/shutdown"
	"github.com/iotaledger/goshimmer/packages/vote"
	"github.com/iotaledger/goshimmer/plugins/messagelayer"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
)

const (
	// PluginName contains the human readable name of the plugin.
	PluginName = "FPCTest"

	// AverageNetworkDelay contains the average time it takes for a network to propagate through gossip.
	AverageNetworkDelay = 6 * time.Second
)

var (
	// App is the "plugin" instance of the value-transfers application.
	App = node.NewPlugin(PluginName, node.Enabled, configure, run)

	// FPCTangle is the FPCTest instance.
	FPCTangle *tangle.Tangle

	// log holds a reference to the logger used by this app.
	log *logger.Logger
)

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)

	log.Debug("configuring FPCTest")

	// create instances
	FPCTangle = tangle.New(database.GetBadgerInstance())

	// subscribe to message-layer
	messagelayer.Tangle.Events.MessageSolid.Attach(events.NewClosure(onReceiveMessageFromMessageLayer))

	// setup behavior of package instances
	FPCTangle.Events.PayloadAttached.Attach(events.NewClosure(onReceiveMessageFromFPCTest))

	configureFPC()
	// TODO: DECIDE WHAT WE SHOULD DO IF FPC FAILS -> cry
	// voter.Events().Failed.Attach(events.NewClosure(panic))

	voter.Events().Finalized.Attach(events.NewClosure(func(id string, opinion vote.Opinion) {
		ID, err := tangle.IDFromBase58(id)
		if err != nil {
			log.Error(err)
			return
		}

		cachedMetadata := FPCTangle.PayloadMetadata(ID)
		defer cachedMetadata.Release()

		metadata := cachedMetadata.Unwrap()

		switch opinion {
		case vote.Like:
			log.Info("Finalized as LIKE: ", ID)
			metadata.SetLike(true)

		case vote.Dislike:
			log.Info("Finalized as DISLIKE: ", ID)
			metadata.SetLike(false)
		}
	}))

	// voter.Events().RoundExecuted.Attach(events.NewClosure(func(stats *vote.RoundStats) {
	// 	log.Info("New Round - ", stats.RandUsed, len(stats.ActiveVoteContexts))
	// }))

	voter.Events().Failed.Attach(events.NewClosure(func(id string, opinion vote.Opinion) {
		log.Info("FPC fail - ", id, opinion)
	}))

}

func run(*node.Plugin) {
	_ = daemon.BackgroundWorker("FPCTangle", func(shutdownSignal <-chan struct{}) {
		<-shutdownSignal
		FPCTangle.Shutdown()
	}, shutdown.PriorityTangle)

	runFPC()
}

func onReceiveMessageFromMessageLayer(cachedMessage *message.CachedMessage, cachedMessageMetadata *messageTangle.CachedMessageMetadata) {
	defer cachedMessage.Release()
	defer cachedMessageMetadata.Release()

	solidMessage := cachedMessage.Unwrap()
	if solidMessage == nil {
		// TODO: LOG ERROR?

		return
	}

	messagePayload := solidMessage.Payload()
	if messagePayload.Type() != fpcTestPayload.Type {
		// TODO: LOG ERROR?

		return
	}

	fpcTestPayload, ok := messagePayload.(*fpcTestPayload.Payload)
	if !ok {
		// TODO: LOG ERROR?

		return
	}

	log.Info("Receive FPCTest Msg - ", fpcTestPayload.ID().String())
	FPCTangle.AttachPayload(fpcTestPayload)
}

func onReceiveMessageFromFPCTest(cachedPayload *fpcTestPayload.CachedPayload, cachedMetadata *tangle.CachedPayloadMetadata) {
	defer cachedPayload.Release()
	defer cachedMetadata.Release()

	log.Info("Conflict detected - ", cachedPayload.Unwrap().ID())

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	like := r.Intn(100)

	switch uint32(like) < cachedPayload.Unwrap().Like() {
	case true:
		err := voter.Vote(cachedPayload.Unwrap().ID().String(), vote.Like)
		if err != nil {
			log.Error(err)
		}
	case false:
		err := voter.Vote(cachedPayload.Unwrap().ID().String(), vote.Dislike)
		if err != nil {
			log.Error(err)
		}
	}

	return
}