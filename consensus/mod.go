package consensus

import (
	"github.com/golang/protobuf/proto"
	"go.dedis.ch/fabric/crypto"
	"go.dedis.ch/fabric/encoding"
	"go.dedis.ch/fabric/mino"
)

// Proposal is the interface that the proposed data must implement to be
// accepted by the consensus.
type Proposal interface {
	encoding.Packable

	GetHash() []byte

	GetPrevious() []byte
}

// Validator is the interface to implement to start a consensus.
type Validator interface {
	Validate(message proto.Message) (Proposal, error)

	Commit(id []byte) error
}

// Participant represents the participant in a consensus.
type Participant interface {
	GetAddress() *mino.Address
	GetPublicKey() crypto.PublicKey
}

// Chain is a verifiable lock between proposals.
type Chain interface {
	encoding.Packable

	Verify(verifier crypto.Verifier, pubkeys []crypto.PublicKey) error
}

// ChainFactory is a factory to decodes chain from protobuf messages.
type ChainFactory interface {
	FromProto(pb proto.Message) (Chain, error)
}

// Consensus is an interface that provides primitives to propose data to a set
// of participants. They will validate the proposal according to the validator.
type Consensus interface {
	GetChainFactory() ChainFactory

	GetChain(from uint64, to uint64) Chain

	Listen(h Validator) error

	Propose(proposal Proposal, nodes ...mino.Node) error
}