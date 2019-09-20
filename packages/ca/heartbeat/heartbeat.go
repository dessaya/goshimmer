package heartbeat

import (
	"sync"

	"github.com/iotaledger/goshimmer/packages/identity"

	"github.com/iotaledger/goshimmer/packages/stringify"

	"github.com/iotaledger/goshimmer/packages/errors"
	"github.com/iotaledger/goshimmer/packages/marshaling"

	"github.com/golang/protobuf/proto"
	heartbeatProto "github.com/iotaledger/goshimmer/packages/ca/heartbeat/proto"
)

type Heartbeat struct {
	nodeId             string
	mainStatement      *OpinionStatement
	neighborStatements map[string][]*OpinionStatement
	signature          []byte

	nodeIdMutex             sync.RWMutex
	mainStatementMutex      sync.RWMutex
	neighborStatementsMutex sync.RWMutex
	signatureMutex          sync.RWMutex
}

func NewHeartbeat() *Heartbeat {
	return &Heartbeat{}
}

func (heartbeat *Heartbeat) GetNodeId() string {
	heartbeat.nodeIdMutex.RLock()
	defer heartbeat.nodeIdMutex.RUnlock()

	return heartbeat.nodeId
}

func (heartbeat *Heartbeat) SetNodeId(nodeId string) {
	heartbeat.nodeIdMutex.Lock()
	defer heartbeat.nodeIdMutex.Unlock()

	heartbeat.nodeId = nodeId
}

func (heartbeat *Heartbeat) GetMainStatement() *OpinionStatement {
	heartbeat.mainStatementMutex.RLock()
	defer heartbeat.mainStatementMutex.RUnlock()

	return heartbeat.mainStatement
}

func (heartbeat *Heartbeat) SetMainStatement(mainStatement *OpinionStatement) {
	heartbeat.mainStatementMutex.Lock()
	defer heartbeat.mainStatementMutex.Unlock()

	heartbeat.mainStatement = mainStatement
}

func (heartbeat *Heartbeat) GetNeighborStatements() map[string][]*OpinionStatement {
	heartbeat.neighborStatementsMutex.RLock()
	defer heartbeat.neighborStatementsMutex.RUnlock()

	return heartbeat.neighborStatements
}

func (heartbeat *Heartbeat) SetNeighborStatements(neighborStatements map[string][]*OpinionStatement) {
	heartbeat.neighborStatementsMutex.Lock()
	defer heartbeat.neighborStatementsMutex.Unlock()

	heartbeat.neighborStatements = neighborStatements
}

func (heartbeat *Heartbeat) GetSignature() []byte {
	heartbeat.signatureMutex.RLock()
	defer heartbeat.signatureMutex.RUnlock()

	return heartbeat.signature
}

func (heartbeat *Heartbeat) SetSignature(signature []byte) {
	heartbeat.signatureMutex.Lock()
	defer heartbeat.signatureMutex.Unlock()

	heartbeat.signature = signature
}

func (heartbeat *Heartbeat) Sign(identity *identity.Identity) (err errors.IdentifiableError) {
	if marshaledHeartbeat, marshalErr := heartbeat.MarshalBinary(); marshalErr == nil {
		if signature, signingErr := identity.Sign(marshaledHeartbeat); signingErr == nil {
			heartbeat.SetSignature(signature)
		} else {
			err = ErrSigningFailed.Derive(signingErr, "failed to sign heartbeat")
		}
	} else {
		err = marshalErr
	}

	return
}

func (heartbeat *Heartbeat) VerifySignature() (result bool, err errors.IdentifiableError) {
	signature := heartbeat.GetSignature()
	heartbeat.SetSignature(nil)

	if marshaledHeartbeat, marshalErr := heartbeat.MarshalBinary(); marshalErr != nil {
		heartbeat.SetSignature(signature)

		err = marshalErr
	} else {
		heartbeat.SetSignature(signature)

		if identity, identityErr := identity.FromSignedData(marshaledHeartbeat, signature); identityErr != nil {
			err = ErrSignatureCorrupt.Derive(identityErr, "failed to retrieve identity from signature of heartbeat")
		} else {
			result = identity.StringIdentifier == heartbeat.GetNodeId()
		}
	}

	return
}

func (heartbeat *Heartbeat) FromProto(proto proto.Message) {
	protoHeartbeat := proto.(*heartbeatProto.HeartBeat)

	var mainStatement OpinionStatement
	mainStatement.FromProto(protoHeartbeat.MainStatement)

	neighborStatements := make(map[string][]*OpinionStatement, len(protoHeartbeat.NeighborStatements))
	for _, neighborStatement := range protoHeartbeat.NeighborStatements {
		var newNeighborStatement OpinionStatement
		newNeighborStatement.FromProto(neighborStatement)

		if _, exists := neighborStatements[neighborStatement.NodeId]; !exists {
			neighborStatements[neighborStatement.NodeId] = make([]*OpinionStatement, 0)
		}

		neighborStatements[neighborStatement.NodeId] = append(neighborStatements[neighborStatement.NodeId], &newNeighborStatement)
	}

	heartbeat.nodeId = protoHeartbeat.NodeId
	heartbeat.mainStatement = &mainStatement
	heartbeat.neighborStatements = neighborStatements
	heartbeat.signature = protoHeartbeat.Signature
}

func (heartbeat *Heartbeat) ToProto() proto.Message {
	neighborStatements := make([]*heartbeatProto.OpinionStatement, len(heartbeat.neighborStatements))
	i := 0
	for _, statementsOfNeighbor := range heartbeat.neighborStatements {
		for _, neighborStatement := range statementsOfNeighbor {
			neighborStatements[i] = neighborStatement.ToProto().(*heartbeatProto.OpinionStatement)

			i++
		}
	}

	return &heartbeatProto.HeartBeat{
		NodeId:             heartbeat.nodeId,
		MainStatement:      heartbeat.mainStatement.ToProto().(*heartbeatProto.OpinionStatement),
		NeighborStatements: neighborStatements,
		Signature:          heartbeat.signature,
	}
}

func (heartbeat *Heartbeat) MarshalBinary() ([]byte, errors.IdentifiableError) {
	return marshaling.Marshal(heartbeat)
}

func (heartbeat *Heartbeat) UnmarshalBinary(data []byte) (err errors.IdentifiableError) {
	return marshaling.Unmarshal(heartbeat, data, &heartbeatProto.HeartBeat{})
}

func (heartbeat *Heartbeat) String() string {
	return stringify.Struct("Heartbeat",
		stringify.StructField("nodeId", heartbeat.nodeId),
		stringify.StructField("mainStatement", heartbeat.mainStatement),
		stringify.StructField("neighborStatements", heartbeat.neighborStatements),
		stringify.StructField("signature", heartbeat.signature),
	)
}
