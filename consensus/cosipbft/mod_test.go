package cosipbft

import (
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/fabric/consensus"
	"go.dedis.ch/fabric/cosi"
	"go.dedis.ch/fabric/cosi/blscosi"
	"go.dedis.ch/fabric/crypto"
	"go.dedis.ch/fabric/encoding"
	"go.dedis.ch/fabric/mino"
	"go.dedis.ch/fabric/mino/minoch"
	"golang.org/x/xerrors"
	nproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/prototest"
)

func TestMessages(t *testing.T) {
	messages := []nproto.Message{
		(*Prepare)(nil),
		(*Commit)(nil),
	}

	for _, m := range messages {
		t.Run(fmt.Sprintf("%T", m), func(t *testing.T) {
			prototest.Message{}.Test(t, m.ProtoReflect().Type())
		})
	}
}

func TestConsensus_Basic(t *testing.T) {
	manager := minoch.NewManager()
	m1, err := minoch.NewMinoch(manager, "A")
	require.NoError(t, err)

	cosi := blscosi.NewBlsCoSi(m1, blscosi.NewSigner())

	cons := NewCoSiPBFT(m1, cosi)
	err = cons.Listen(fakeValidator{pubkey: cosi.GetPublicKey()})
	require.NoError(t, err)

	prop := fakeProposal{}
	err = cons.Propose(prop, fakeParticipant{
		addr:      m1.GetAddress(),
		publicKey: cosi.GetPublicKey(),
	})
	require.NoError(t, err)

	ch, err := cons.GetChain(prop.GetHash())
	require.NoError(t, err)
	require.Len(t, ch.(forwardLinkChain).links, 1)
}

func TestConsensus_GetChainFactory(t *testing.T) {
	factory := &defaultChainFactory{}
	cons := &Consensus{
		factory: factory,
	}

	require.Equal(t, factory, cons.GetChainFactory())
}

type fakeStorage struct {
	Storage
}

func (s fakeStorage) Store(*ForwardLinkProto) error {
	return xerrors.New("oops")
}

func (s fakeStorage) ReadChain(id Digest) ([]*ForwardLinkProto, error) {
	return nil, xerrors.New("oops")
}

func (s fakeStorage) ReadLast() (*ForwardLinkProto, error) {
	return nil, xerrors.New("oops")
}

func TestConsensus_GetChain(t *testing.T) {
	cons := &Consensus{
		storage: newInMemoryStorage(),
		factory: &defaultChainFactory{verifier: fakeVerifierWithFactory{}},
	}

	err := cons.storage.Store(&ForwardLinkProto{To: []byte{0xaa}})
	require.NoError(t, err)

	chain, err := cons.GetChain([]byte{0xaa})
	require.NoError(t, err)
	require.Len(t, chain.(forwardLinkChain).links, 1)

	err = cons.storage.Store(&ForwardLinkProto{From: []byte{0xaa}, To: []byte{0xbb}, Prepare: &any.Any{}})
	require.NoError(t, err)
	_, err = cons.GetChain([]byte{0xbb})
	require.EqualError(t, err,
		"couldn't decode chain: couldn't decode forward link: couldn't decode prepare signature: oops")

	cons.storage = fakeStorage{}
	_, err = cons.GetChain([]byte{})
	require.EqualError(t, err, "couldn't read the chain: oops")
}

func TestConsensus_Listen(t *testing.T) {
	fakeCosi := &fakeCosi{}
	fakeMino := &fakeMino{}
	cons := &Consensus{cosi: fakeCosi, mino: fakeMino}

	err := cons.Listen(fakeValidator{})
	require.NoError(t, err)
	require.IsType(t, handler{}, fakeCosi.handler)
	require.IsType(t, rpcHandler{}, fakeMino.h)
	require.Equal(t, rpcName, fakeMino.name)

	err = cons.Listen(nil)
	require.EqualError(t, err, "validator is nil")

	fakeCosi.err = xerrors.New("cosi error")
	err = cons.Listen(fakeValidator{})
	require.Error(t, err)
	require.True(t, xerrors.Is(err, fakeCosi.err))

	fakeCosi.err = nil
	fakeMino.err = xerrors.New("rpc error")
	err = cons.Listen(fakeValidator{})
	require.Error(t, err)
	require.True(t, xerrors.Is(err, fakeMino.err))
}

func checkSignatureValue(t *testing.T, pb *any.Any, value uint64) {
	var wrapper wrappers.UInt64Value
	err := ptypes.UnmarshalAny(pb, &wrapper)
	require.NoError(t, err)
	require.Equal(t, value, wrapper.GetValue())
}

func TestConsensus_Propose(t *testing.T) {
	cs := &fakeCosi{}
	rpc := &fakeRPC{}
	cons := &Consensus{cosi: cs, rpc: rpc}

	err := cons.Propose(fakeProposal{}, fakeParticipant{})
	require.NoError(t, err)
	require.Len(t, cs.calls, 2)

	prepare := cs.calls[0]["message"].(*Prepare)
	require.NotNil(t, prepare)
	signers := cs.calls[0]["signers"].([]cosi.Cosigner)
	require.Len(t, signers, 1)

	commit := cs.calls[1]["message"].(*Commit)
	require.NotNil(t, commit)
	require.Equal(t, []byte{0xaa}, commit.GetTo())
	checkSignatureValue(t, commit.GetPrepare(), 1)
	signers = cs.calls[1]["signers"].([]cosi.Cosigner)
	require.Len(t, signers, 1)

	require.Len(t, rpc.calls, 1)
	propagate := rpc.calls[0]["message"].(*Propagate)
	require.NotNil(t, propagate)
	require.Equal(t, []byte{0xaa}, propagate.GetTo())
	checkSignatureValue(t, propagate.GetCommit(), 2)
}

type badCosi struct {
	cosi.CollectiveSigning
	delay int
}

func (cs *badCosi) Sign(pb proto.Message, signers ...cosi.Cosigner) (crypto.Signature, error) {
	if cs.delay > 0 {
		cs.delay--
		return fakeSignature{}, nil
	}

	return fakeSignature{err: xerrors.New("oops")}, nil
}

func TestConsensus_ProposeFailures(t *testing.T) {
	defer func() { protoenc = encoding.NewProtoEncoder() }()

	cons := &Consensus{}

	e := xerrors.New("pack error")
	err := cons.Propose(fakeProposal{err: e})
	require.Error(t, err)
	require.True(t, xerrors.Is(err, encoding.NewEncodingError("proposal", nil)))

	protoenc = &fakeEncoder{}
	err = cons.Propose(fakeProposal{})
	require.Error(t, err)
	require.True(t, xerrors.Is(err,
		encoding.NewAnyEncodingError((*empty.Empty)(nil), nil)))

	protoenc = encoding.NewProtoEncoder()
	err = cons.Propose(fakeProposal{}, fakeNode{})
	require.EqualError(t, err, "node must implement cosi.Cosigner")

	cons.cosi = &fakeCosi{err: xerrors.New("oops")}
	err = cons.Propose(fakeProposal{})
	require.EqualError(t, err, "couldn't sign the proposal: oops")

	cons.cosi = &badCosi{}
	err = cons.Propose(fakeProposal{})
	require.EqualError(t, err, "couldn't encode prepare signature: oops")

	cons.cosi = &badCosi{delay: 1}
	protoenc = &fakeEncoder{delay: 1}
	err = cons.Propose(fakeProposal{})
	require.EqualError(t, xerrors.Unwrap(err), "marshal any error")

	cons.cosi = &fakeCosi{err: xerrors.New("oops"), delay: 1}
	protoenc = encoding.NewProtoEncoder()
	err = cons.Propose(fakeProposal{})
	require.EqualError(t, err, "couldn't sign the commit: oops")

	cons.cosi = &badCosi{delay: 1}
	err = cons.Propose(fakeProposal{})
	require.EqualError(t, err, "couldn't encode commit signature: oops")

	cons.cosi = &badCosi{delay: 2}
	protoenc = &fakeEncoder{delay: 2}
	err = cons.Propose(fakeProposal{})
	require.EqualError(t, xerrors.Unwrap(err), "marshal any error")

	protoenc = encoding.NewProtoEncoder()
	cons.cosi = &fakeCosi{}
	cons.rpc = &fakeRPC{err: xerrors.New("oops")}
	err = cons.Propose(fakeProposal{})
	require.EqualError(t, err, "couldn't propagate the link: oops")
}

func TestHandler_HashPrepare(t *testing.T) {
	h := handler{
		validator: fakeValidator{},
		Consensus: &Consensus{
			storage: newInMemoryStorage(),
			queue:   &queue{},
		},
	}

	_, err := h.Hash(&empty.Empty{})
	require.EqualError(t, err, "message type not supported: *emptypb.Empty")

	empty, err := ptypes.MarshalAny(&empty.Empty{})
	require.NoError(t, err)

	buffer, err := h.Hash(&Prepare{Proposal: empty})
	require.NoError(t, err)
	require.NotEmpty(t, buffer)

	_, err = h.Hash(&Prepare{Proposal: nil})
	require.Error(t, err)
	require.True(t, xerrors.Is(err, encoding.NewAnyDecodingError((*ptypes.DynamicAny)(nil), nil)))

	h.validator = fakeValidator{err: xerrors.New("oops")}
	_, err = h.Hash(&Prepare{Proposal: empty})
	require.EqualError(t, err, "couldn't validate the proposal: oops")

	h.validator = fakeValidator{}
	h.Consensus = &Consensus{storage: fakeStorage{}}
	_, err = h.Hash(&Prepare{Proposal: empty})
	require.EqualError(t, err, "couldn't read last: oops")

	h.Consensus.storage = newInMemoryStorage()
	h.Consensus.storage.Store(&ForwardLinkProto{To: []byte{0xbb}})
	_, err = h.Hash(&Prepare{Proposal: empty})
	require.EqualError(t, err, "mismatch with previous link: bb != aa")

	h.Consensus.storage = newInMemoryStorage()
	h.Consensus.queue = &queue{locked: true}
	_, err = h.Hash(&Prepare{Proposal: empty})
	require.EqualError(t, err, "couldn't add to queue: queue is locked")
}

type fakeFactory struct {
	ChainFactory
	err          error
	errSignature error
}

func (f fakeFactory) DecodeSignature(pb proto.Message) (crypto.Signature, error) {
	return fakeSignature{err: f.errSignature}, f.err
}

func (f fakeFactory) FromProto(pb proto.Message) (consensus.Chain, error) {
	return forwardLinkChain{}, f.err
}

func TestHandler_HashCommit(t *testing.T) {
	h := handler{
		Consensus: &Consensus{
			factory: fakeFactory{},
			storage: newInMemoryStorage(),
			queue:   &queue{verifier: &fakeVerifier{}},
		},
	}

	err := h.Consensus.queue.New(fakeProposal{}, fakeProposal{})
	require.NoError(t, err)

	buffer, err := h.Hash(&Commit{To: []byte{0xaa}})
	require.NoError(t, err)
	require.Equal(t, []byte{0xde, 0xad, 0xbe, 0xef}, buffer)

	h.Consensus.factory = fakeFactory{err: xerrors.New("oops")}
	_, err = h.Hash(&Commit{})
	require.EqualError(t, err, "couldn't decode prepare signature: oops")

	h.Consensus.factory = fakeFactory{}
	_, err = h.Hash(&Commit{To: []byte("unknown")})
	require.EqualError(t, err, "couldn't update signature: couldn't find proposal '756e6b6e6f776e'")

	h.Consensus.factory = fakeFactory{errSignature: xerrors.New("oops")}
	_, err = h.Hash(&Commit{To: []byte{0xaa}})
	require.EqualError(t, err, "couldn't marshal the signature: oops")
}

type fakeQueue struct {
	Queue
	err error
}

func (q fakeQueue) Finalize(to Digest, commit crypto.Signature) (*ForwardLinkProto, error) {
	return &ForwardLinkProto{}, q.err
}

func TestRPCHandler_Process(t *testing.T) {
	h := rpcHandler{
		validator: fakeValidator{},
		Consensus: &Consensus{
			factory: fakeFactory{},
			storage: newInMemoryStorage(),
			queue:   fakeQueue{},
		},
	}

	resp, err := h.Process(&empty.Empty{})
	require.EqualError(t, err, "message type not supported: *emptypb.Empty")
	require.Nil(t, resp)

	resp, err = h.Process(&Propagate{})
	require.NoError(t, err)
	require.Nil(t, resp)

	h.Consensus.factory = fakeFactory{err: xerrors.New("oops")}
	_, err = h.Process(&Propagate{})
	require.EqualError(t, err, "couldn't decode commit signature: oops")

	h.Consensus.factory = fakeFactory{}
	h.Consensus.queue = fakeQueue{err: xerrors.New("oops")}
	_, err = h.Process(&Propagate{})
	require.EqualError(t, err, "couldn't finalize: oops")

	h.Consensus.queue = fakeQueue{}
	h.Consensus.storage = fakeStorage{}
	_, err = h.Process(&Propagate{})
	require.EqualError(t, err, "couldn't write forward link: oops")

	h.Consensus.storage = newInMemoryStorage()
	h.validator = fakeValidator{err: xerrors.New("oops")}
	_, err = h.Process(&Propagate{})
	require.EqualError(t, err, "couldn't commit: oops")
}

type fakeProposal struct {
	pubkeys []crypto.PublicKey
	err     error
}

func (p fakeProposal) Pack() (proto.Message, error) {
	return &empty.Empty{}, p.err
}

func (p fakeProposal) GetHash() []byte {
	return []byte{0xaa}
}

func (p fakeProposal) GetPublicKeys() []crypto.PublicKey {
	return p.pubkeys
}

type fakeValidator struct {
	pubkey crypto.PublicKey
	err    error
}

func (v fakeValidator) Validate(msg proto.Message) (consensus.Proposal, consensus.Proposal, error) {
	p := fakeProposal{
		pubkeys: []crypto.PublicKey{v.pubkey},
	}
	return p, p, v.err
}

func (v fakeValidator) Commit(id []byte) error {
	return v.err
}

type fakeNode struct {
	mino.Node
}

type fakeParticipant struct {
	addr      mino.Address
	publicKey crypto.PublicKey
}

func (p fakeParticipant) GetAddress() mino.Address {
	return p.addr
}

func (p fakeParticipant) GetPublicKey() crypto.PublicKey {
	return p.publicKey
}

type fakeSignature struct {
	crypto.Signature
	value uint64
	err   error
}

func (s fakeSignature) Pack() (proto.Message, error) {
	return &wrappers.UInt64Value{Value: s.value}, s.err
}

func (s fakeSignature) MarshalBinary() ([]byte, error) {
	return []byte{0xde, 0xad, 0xbe, 0xef}, s.err
}

type fakeCosi struct {
	cosi.CollectiveSigning
	handler cosi.Hashable
	calls   []map[string]interface{}
	count   uint64
	delay   int
	err     error
}

func (cs *fakeCosi) Listen(h cosi.Hashable) error {
	cs.handler = h
	return cs.err
}

func (cs *fakeCosi) Sign(pb proto.Message, signers ...cosi.Cosigner) (crypto.Signature, error) {
	// Increase the counter each time a test sign a message.
	cs.count++
	// Store the call parameters so that they can be verified in the test.
	cs.calls = append(cs.calls, map[string]interface{}{
		"message": pb,
		"signers": signers,
	})
	if cs.err != nil {
		if cs.delay == 0 {
			return nil, cs.err
		}
		cs.delay--
	}
	return fakeSignature{value: cs.count}, nil
}

type fakeRPC struct {
	mino.RPC

	calls []map[string]interface{}
	err   error
}

func (rpc *fakeRPC) Call(pb proto.Message, nodes ...mino.Node) (<-chan proto.Message, <-chan error) {
	msgs := make(chan proto.Message, 0)
	errs := make(chan error, 1)
	if rpc.err != nil {
		errs <- rpc.err
	} else {
		close(msgs)
	}
	rpc.calls = append(rpc.calls, map[string]interface{}{
		"message": pb,
		"nodes":   nodes,
	})
	return msgs, errs
}

type fakeMino struct {
	mino.Mino
	name string
	h    mino.Handler
	err  error
}

func (m *fakeMino) MakeRPC(name string, h mino.Handler) (mino.RPC, error) {
	m.name = name
	m.h = h
	return nil, m.err
}
