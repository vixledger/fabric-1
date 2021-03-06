package cosipbft

import (
	"bytes"
	"sync"

	"go.dedis.ch/fabric/consensus"
	"go.dedis.ch/fabric/crypto"
	"go.dedis.ch/fabric/encoding"
	"golang.org/x/xerrors"
)

// Queue is an interface specific to cosipbft that defines the primitives to
// prepare and commit to proposals.
type Queue interface {
	New(curr consensus.Proposal) error
	LockProposal(to Digest, sig crypto.Signature) error
	Finalize(to Digest, sig crypto.Signature) (*ForwardLinkProto, error)
}

type item struct {
	from     Digest
	to       Digest
	prepare  crypto.Signature
	verifier crypto.Verifier
}

type queue struct {
	sync.Mutex
	locked       bool
	chainFactory ChainFactory
	items        []item
}

func newQueue(factory ChainFactory) *queue {
	return &queue{
		locked:       false,
		chainFactory: factory,
	}
}

func (q *queue) getItem(id Digest) (item, int, bool) {
	for i, item := range q.items {
		if bytes.Equal(id, item.to) {
			return item, i, true
		}
	}

	return item{}, -1, false
}

func (q *queue) New(curr consensus.Proposal) error {
	q.Lock()
	defer q.Unlock()

	if q.locked {
		return xerrors.New("queue is locked")
	}

	_, _, ok := q.getItem(curr.GetHash())
	if ok {
		return xerrors.Errorf("proposal '%x' already exists", curr.GetHash())
	}

	q.items = append(q.items, item{
		to:       curr.GetHash(),
		from:     curr.GetPreviousHash(),
		verifier: curr.GetVerifier(),
	})
	return nil
}

// LockProposal verifies the prepare signature and stores it. It also locks
// the queue to prevent further committing.
func (q *queue) LockProposal(to Digest, sig crypto.Signature) error {
	q.Lock()
	defer q.Unlock()

	if q.locked {
		return xerrors.New("queue is locked")
	}

	item, index, ok := q.getItem(to)
	if !ok {
		return xerrors.Errorf("couldn't find proposal '%x'", to)
	}

	forwardLink := forwardLink{
		from: item.from,
		to:   item.to,
	}

	hash, err := forwardLink.computeHash(q.chainFactory.GetHashFactory().New())
	if err != nil {
		return xerrors.Errorf("couldn't hash proposal: %v", err)
	}

	err = item.verifier.Verify(hash, sig)
	if err != nil {
		return xerrors.Errorf("couldn't verify signature: %v", err)
	}

	q.items[index].prepare = sig
	q.locked = true

	return nil
}

// Finalize verifies the commit signature and clear the queue.
func (q *queue) Finalize(to Digest, sig crypto.Signature) (*ForwardLinkProto, error) {
	q.Lock()
	defer q.Unlock()

	item, _, ok := q.getItem(to)
	if !ok {
		return nil, xerrors.Errorf("couldn't find proposal '%x'", to)
	}

	if item.prepare == nil {
		return nil, xerrors.Errorf("no signature for proposal '%x'", to)
	}

	forwardLink := forwardLink{
		from:    item.from,
		to:      item.to,
		prepare: item.prepare,
		commit:  sig,
	}

	// Make sure the commit signature is a valid one before committing.
	buffer, err := forwardLink.prepare.MarshalBinary()
	if err != nil {
		return nil, xerrors.Errorf("couldn't marshal the signature: %v", err)
	}

	err = item.verifier.Verify(buffer, sig)
	if err != nil {
		return nil, xerrors.Errorf("couldn't verify signature: %v", err)
	}

	packed, err := forwardLink.Pack()
	if err != nil {
		return nil, encoding.NewEncodingError("forward link", err)
	}

	q.locked = false
	q.items = nil

	return packed.(*ForwardLinkProto), nil
}
