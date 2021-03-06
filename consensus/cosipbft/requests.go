package cosipbft

import (
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"go.dedis.ch/fabric/consensus"
	"go.dedis.ch/fabric/crypto"
	"go.dedis.ch/fabric/encoding"
	"golang.org/x/xerrors"
)

// Prepare is the request sent at the beginning of the PBFT protocol.
type Prepare struct {
	proposal consensus.Proposal
	digest   []byte
}

func newPrepareRequest(prop consensus.Proposal, f crypto.HashFactory) (Prepare, error) {
	forwardLink := forwardLink{
		from: prop.GetPreviousHash(),
		to:   prop.GetHash(),
	}

	req := Prepare{proposal: prop}

	var err error
	req.digest, err = forwardLink.computeHash(f.New())
	if err != nil {
		return req, xerrors.Errorf("couldn't compute hash: %v", err)
	}

	return req, nil
}

// GetHash returns the hash of the prepare request that will be signed by the
// collective authority.
func (p Prepare) GetHash() []byte {
	return p.digest
}

// Pack returns the protobuf message, or an error.
func (p Prepare) Pack() (proto.Message, error) {
	packed, err := p.proposal.Pack()
	if err != nil {
		return nil, encoding.NewEncodingError("proposal", err)
	}

	pb := &PrepareRequest{}
	pb.Proposal, err = ptypes.MarshalAny(packed)
	if err != nil {
		return nil, encoding.NewAnyEncodingError(packed, err)
	}

	return pb, nil
}

// Commit is the request sent for the last phase of the PBFT.
type Commit struct {
	to      Digest
	prepare crypto.Signature
	hash    []byte
}

func newCommitRequest(to []byte, prepare crypto.Signature) (Commit, error) {
	buffer, err := prepare.MarshalBinary()
	if err != nil {
		return Commit{}, xerrors.Errorf("couldn't marshal prepare signature: %v", err)
	}

	commit := Commit{
		to:      to,
		prepare: prepare,
		hash:    buffer,
	}

	return commit, nil
}

// GetHash returns the hash for the commit message. The actual value is the
// marshaled prepare signature.
func (c Commit) GetHash() []byte {
	return c.hash
}

// Pack returns the protobuf message representation of a commit, or an error if
// something goes wrong during encoding.
func (c Commit) Pack() (proto.Message, error) {
	packed, err := c.prepare.Pack()
	if err != nil {
		return nil, encoding.NewEncodingError("prepare signature", err)
	}

	packedany, err := ptypes.MarshalAny(packed)
	if err != nil {
		return nil, encoding.NewAnyEncodingError(packed, err)
	}

	pb := &CommitRequest{
		To:      c.to,
		Prepare: packedany,
	}

	return pb, nil
}
