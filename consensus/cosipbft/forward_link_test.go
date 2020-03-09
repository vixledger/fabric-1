package cosipbft

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	any "github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/fabric/crypto"
	"go.dedis.ch/fabric/encoding"
	"golang.org/x/xerrors"
)

func TestForwardLink_Verify(t *testing.T) {
	fl := forwardLink{
		hash:    []byte{0xaa},
		prepare: fakeSignature{},
	}

	verifier := &fakeVerifier{}
	pubkeys := []crypto.PublicKey{fakePublicKey{}}

	err := fl.Verify(verifier, pubkeys)
	require.NoError(t, err)
	require.Len(t, verifier.calls, 2)
	require.Equal(t, []byte{0xaa}, verifier.calls[0]["message"])
	require.Equal(t, pubkeys, verifier.calls[0]["pubkeys"])
	require.Equal(t, []byte{0xde, 0xad, 0xbe, 0xef}, verifier.calls[1]["message"])
	require.Equal(t, pubkeys, verifier.calls[1]["pubkeys"])

	verifier.err = xerrors.New("oops")
	err = fl.Verify(verifier, nil)
	require.EqualError(t, err, "couldn't verify prepare signature: oops")

	verifier.delay = 1
	err = fl.Verify(verifier, nil)
	require.EqualError(t, err, "couldn't verify commit signature: oops")

	verifier.err = nil
	fl.prepare = fakeSignature{err: xerrors.New("oops")}
	err = fl.Verify(verifier, pubkeys)
	require.EqualError(t, err, "couldn't marshal the signature: oops")
}

func TestForwardLink_Pack(t *testing.T) {
	defer func() { protoenc = encoding.NewProtoEncoder() }()

	fl := forwardLink{
		from: []byte{0xaa},
		to:   []byte{0xbb},
	}

	pb, err := fl.Pack()
	require.NoError(t, err)
	flp, ok := pb.(*ForwardLinkProto)
	require.True(t, ok)
	require.Equal(t, flp.GetFrom(), fl.from)
	require.Equal(t, flp.GetTo(), fl.to)
	require.Nil(t, flp.GetPrepare())
	require.Nil(t, flp.GetCommit())

	fl.prepare = fakeSignature{value: 1}
	pb, err = fl.Pack()
	require.NoError(t, err)
	flp = pb.(*ForwardLinkProto)
	checkSignatureValue(t, flp.GetPrepare(), 1)

	fl.commit = fakeSignature{value: 2}
	pb, err = fl.Pack()
	require.NoError(t, err)
	flp = pb.(*ForwardLinkProto)
	checkSignatureValue(t, flp.GetCommit(), 2)

	// Test if the prepare signature cannot be packed.
	fl.prepare = fakeSignature{err: xerrors.New("oops")}
	pb, err = fl.Pack()
	require.EqualError(t, xerrors.Unwrap(err), "couldn't pack: oops")

	// Test if the commit signature cannot be packed.
	fl.prepare = nil
	fl.commit = fakeSignature{}
	protoenc = &fakeEncoder{}
	pb, err = fl.Pack()
	require.EqualError(t, xerrors.Unwrap(xerrors.Unwrap(err)), "marshal any error")
}

func TestChain_Verify(t *testing.T) {
	pubkeys := []crypto.PublicKey{fakePublicKey{}}
	chain := forwardLinkChain{
		links: []forwardLink{
			{from: []byte{0xaa}, to: []byte{0xbb}, prepare: fakeSignature{}, commit: fakeSignature{}},
			{from: []byte{0xbb}, to: []byte{0xcc}, prepare: fakeSignature{}, commit: fakeSignature{}},
		},
	}

	verifier := &fakeVerifier{}
	err := chain.Verify(verifier, pubkeys)
	require.NoError(t, err)
	require.Len(t, verifier.calls, 4)

	err = chain.Verify(&fakeVerifier{err: xerrors.New("oops")}, pubkeys)
	require.EqualError(t, xerrors.Unwrap(err), "couldn't verify prepare signature: oops")

	chain.links[0].to = []byte{0xff}
	err = chain.Verify(&fakeVerifier{}, pubkeys)
	require.EqualError(t, err, "mismatch forward link 'ff' != 'bb'")

	chain.links = nil
	err = chain.Verify(&fakeVerifier{}, pubkeys)
	require.EqualError(t, err, "chain is empty")
}

func TestChain_Pack(t *testing.T) {
	chain := forwardLinkChain{
		links: []forwardLink{
			{},
			{},
		},
	}

	pb, err := chain.Pack()
	require.NoError(t, err)
	require.IsType(t, (*ChainProto)(nil), pb)
	require.Len(t, pb.(*ChainProto).GetLinks(), 2)

	chain.links[0].prepare = fakeSignature{err: xerrors.New("oops")}
	pb, err = chain.Pack()
	require.EqualError(t, err, "couldn't encode forward link: "+
		"couldn't encode prepare: couldn't pack: oops")
}

type fakeSignatureFactory struct {
	crypto.SignatureFactory
}

func (f fakeSignatureFactory) FromProto(pb proto.Message) (crypto.Signature, error) {
	return nil, xerrors.New("oops")
}

type fakeVerifierWithFactory struct {
	crypto.Verifier
}

func (v fakeVerifierWithFactory) GetSignatureFactory() crypto.SignatureFactory {
	return fakeSignatureFactory{}
}

func TestChainFactory_FromProto(t *testing.T) {
	defer func() { protoenc = encoding.NewProtoEncoder() }()

	chainpb := &ChainProto{
		Links: []*ForwardLinkProto{
			{},
			{},
		},
	}

	factory := newChainFactory(fakeVerifierWithFactory{})
	chain, err := factory.FromProto(chainpb)
	require.NoError(t, err)
	require.NotNil(t, chain)

	chainany, err := ptypes.MarshalAny(chainpb)
	require.NoError(t, err)

	chain, err = factory.FromProto(chainany)
	require.NoError(t, err)
	require.NotNil(t, chain)

	_, err = factory.FromProto(&empty.Empty{})
	require.EqualError(t, err, "message type not supported: *emptypb.Empty")

	chainpb.Links[0].Prepare = &any.Any{}
	_, err = factory.FromProto(chainpb)
	require.EqualError(t, xerrors.Unwrap(err), "couldn't decode prepare signature: oops")

	protoenc = &fakeEncoder{}
	_, err = factory.FromProto(chainany)
	require.EqualError(t, err, "couldn't decode *cosipbft.ChainProto to any: unmarshal any error")
}

func TestChainFactory_DecodeSignature(t *testing.T) {
	factory := defaultChainFactory{verifier: fakeVerifierWithFactory{}}
	_, err := factory.DecodeSignature(nil)
	require.EqualError(t, err, "oops")
}

func TestChainFactory_DecodeForwardLink(t *testing.T) {
	defer func() { protoenc = encoding.NewProtoEncoder() }()

	factory := defaultChainFactory{}

	forwardLink := &ForwardLinkProto{}
	flany, err := ptypes.MarshalAny(forwardLink)
	require.NoError(t, err)

	chain, err := factory.DecodeForwardLink(flany)
	require.NoError(t, err)
	require.NotNil(t, chain)

	_, err = factory.DecodeForwardLink(&empty.Empty{})
	require.EqualError(t, err, "unknown message type: *emptypb.Empty")

	factory.verifier = fakeVerifierWithFactory{}
	forwardLink.Prepare = &any.Any{}
	_, err = factory.DecodeForwardLink(forwardLink)
	require.EqualError(t, err, "couldn't decode prepare signature: oops")

	forwardLink.Prepare = nil
	forwardLink.Commit = &any.Any{}
	_, err = factory.DecodeForwardLink(forwardLink)
	require.EqualError(t, err, "couldn't decode commit signature: oops")

	protoenc = &fakeEncoder{}
	_, err = factory.DecodeForwardLink(flany)
	require.Error(t, err)
	require.True(t, xerrors.Is(err, encoding.NewAnyDecodingError((*ForwardLinkProto)(nil), nil)))
}

type fakePublicKey struct {
	crypto.PublicKey
}

type fakeEncoder struct {
	encoding.ProtoMarshaler
	delay int
}

func (e *fakeEncoder) MarshalAny(pb proto.Message) (*any.Any, error) {
	if e.delay > 0 {
		e.delay--
		return ptypes.MarshalAny(pb)
	}
	return nil, xerrors.New("marshal any error")
}

func (e *fakeEncoder) UnmarshalAny(*any.Any, proto.Message) error {
	return xerrors.New("unmarshal any error")
}
