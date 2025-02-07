package crypto

import (
	"bytes"

	secp256k1 "github.com/btcsuite/btcd/btcec"
	"github.com/tepleton/ed25519"
	"github.com/tepleton/ed25519/extra25519"
	. "github.com/tepleton/go-common"
	data "github.com/tepleton/go-data"
	"github.com/tepleton/go-wire"
)

func PrivKeyFromBytes(privKeyBytes []byte) (privKey PrivKey, err error) {
	err = wire.ReadBinaryBytes(privKeyBytes, &privKey)
	return
}

//----------------------------------------

type PrivKey struct {
	PrivKeyInner `json:"unwrap"`
}

// DO NOT USE THIS INTERFACE.
// You probably want to use PubKey
type PrivKeyInner interface {
	AssertIsPrivKeyInner()
	Bytes() []byte
	Sign(msg []byte) Signature
	PubKey() PubKey
	Equals(PrivKey) bool
	Wrap() PrivKey
}

func (p PrivKey) MarshalJSON() ([]byte, error) {
	return privKeyMapper.ToJSON(p.PrivKeyInner)
}

func (p *PrivKey) UnmarshalJSON(data []byte) (err error) {
	parsed, err := privKeyMapper.FromJSON(data)
	if err == nil && parsed != nil {
		p.PrivKeyInner = parsed.(PrivKeyInner)
	}
	return
}

// Unwrap recovers the concrete interface safely (regardless of levels of embeds)
func (p PrivKey) Unwrap() PrivKeyInner {
	pk := p.PrivKeyInner
	for wrap, ok := pk.(PrivKey); ok; wrap, ok = pk.(PrivKey) {
		pk = wrap.PrivKeyInner
	}
	return pk
}

func (p PrivKey) Empty() bool {
	return p.PrivKeyInner == nil
}

var privKeyMapper = data.NewMapper(PrivKey{}).
	RegisterImplementation(PrivKeyEd25519{}, NameEd25519, TypeEd25519).
	RegisterImplementation(PrivKeySecp256k1{}, NameSecp256k1, TypeSecp256k1)

//-------------------------------------

var _ PrivKeyInner = PrivKeyEd25519{}

// Implements PrivKey
type PrivKeyEd25519 [64]byte

func (privKey PrivKeyEd25519) AssertIsPrivKeyInner() {}

func (privKey PrivKeyEd25519) Bytes() []byte {
	return wire.BinaryBytes(PrivKey{privKey})
}

func (privKey PrivKeyEd25519) Sign(msg []byte) Signature {
	privKeyBytes := [64]byte(privKey)
	signatureBytes := ed25519.Sign(&privKeyBytes, msg)
	return SignatureEd25519(*signatureBytes).Wrap()
}

func (privKey PrivKeyEd25519) PubKey() PubKey {
	privKeyBytes := [64]byte(privKey)
	pubBytes := *ed25519.MakePublicKey(&privKeyBytes)
	return PubKeyEd25519(pubBytes).Wrap()
}

func (privKey PrivKeyEd25519) Equals(other PrivKey) bool {
	if otherEd, ok := other.Unwrap().(PrivKeyEd25519); ok {
		return bytes.Equal(privKey[:], otherEd[:])
	} else {
		return false
	}
}

func (p PrivKeyEd25519) MarshalJSON() ([]byte, error) {
	return data.Encoder.Marshal(p[:])
}

func (p *PrivKeyEd25519) UnmarshalJSON(enc []byte) error {
	var ref []byte
	err := data.Encoder.Unmarshal(&ref, enc)
	copy(p[:], ref)
	return err
}

func (privKey PrivKeyEd25519) ToCurve25519() *[32]byte {
	keyCurve25519 := new([32]byte)
	privKeyBytes := [64]byte(privKey)
	extra25519.PrivateKeyToCurve25519(keyCurve25519, &privKeyBytes)
	return keyCurve25519
}

func (privKey PrivKeyEd25519) String() string {
	return Fmt("PrivKeyEd25519{*****}")
}

// Deterministically generates new priv-key bytes from key.
func (privKey PrivKeyEd25519) Generate(index int) PrivKeyEd25519 {
	newBytes := wire.BinarySha256(struct {
		PrivKey [64]byte
		Index   int
	}{privKey, index})
	var newKey [64]byte
	copy(newKey[:], newBytes)
	return PrivKeyEd25519(newKey)
}

func (privKey PrivKeyEd25519) Wrap() PrivKey {
	return PrivKey{privKey}
}

func GenPrivKeyEd25519() PrivKeyEd25519 {
	privKeyBytes := new([64]byte)
	copy(privKeyBytes[:32], CRandBytes(32))
	ed25519.MakePublicKey(privKeyBytes)
	return PrivKeyEd25519(*privKeyBytes)
}

// NOTE: secret should be the output of a KDF like bcrypt,
// if it's derived from user input.
func GenPrivKeyEd25519FromSecret(secret []byte) PrivKeyEd25519 {
	privKey32 := Sha256(secret) // Not Ripemd160 because we want 32 bytes.
	privKeyBytes := new([64]byte)
	copy(privKeyBytes[:32], privKey32)
	ed25519.MakePublicKey(privKeyBytes)
	return PrivKeyEd25519(*privKeyBytes)
}

//-------------------------------------

var _ PrivKeyInner = PrivKeySecp256k1{}

// Implements PrivKey
type PrivKeySecp256k1 [32]byte

func (privKey PrivKeySecp256k1) AssertIsPrivKeyInner() {}

func (privKey PrivKeySecp256k1) Bytes() []byte {
	return wire.BinaryBytes(PrivKey{privKey})
}

func (privKey PrivKeySecp256k1) Sign(msg []byte) Signature {
	priv__, _ := secp256k1.PrivKeyFromBytes(secp256k1.S256(), privKey[:])
	sig__, err := priv__.Sign(Sha256(msg))
	if err != nil {
		PanicSanity(err)
	}
	return SignatureSecp256k1(sig__.Serialize()).Wrap()
}

func (privKey PrivKeySecp256k1) PubKey() PubKey {
	_, pub__ := secp256k1.PrivKeyFromBytes(secp256k1.S256(), privKey[:])
	var pub PubKeySecp256k1
	copy(pub[:], pub__.SerializeCompressed())
	return pub.Wrap()
}

func (privKey PrivKeySecp256k1) Equals(other PrivKey) bool {
	if otherSecp, ok := other.Unwrap().(PrivKeySecp256k1); ok {
		return bytes.Equal(privKey[:], otherSecp[:])
	} else {
		return false
	}
}

func (p PrivKeySecp256k1) MarshalJSON() ([]byte, error) {
	return data.Encoder.Marshal(p[:])
}

func (p *PrivKeySecp256k1) UnmarshalJSON(enc []byte) error {
	var ref []byte
	err := data.Encoder.Unmarshal(&ref, enc)
	copy(p[:], ref)
	return err
}

func (privKey PrivKeySecp256k1) String() string {
	return Fmt("PrivKeySecp256k1{*****}")
}

func (privKey PrivKeySecp256k1) Wrap() PrivKey {
	return PrivKey{privKey}
}

/*
// Deterministically generates new priv-key bytes from key.
func (key PrivKeySecp256k1) Generate(index int) PrivKeySecp256k1 {
	newBytes := wire.BinarySha256(struct {
		PrivKey [64]byte
		Index   int
	}{key, index})
	var newKey [64]byte
	copy(newKey[:], newBytes)
	return PrivKeySecp256k1(newKey)
}
*/

func GenPrivKeySecp256k1() PrivKeySecp256k1 {
	privKeyBytes := [32]byte{}
	copy(privKeyBytes[:], CRandBytes(32))
	priv, _ := secp256k1.PrivKeyFromBytes(secp256k1.S256(), privKeyBytes[:])
	copy(privKeyBytes[:], priv.Serialize())
	return PrivKeySecp256k1(privKeyBytes)
}

// NOTE: secret should be the output of a KDF like bcrypt,
// if it's derived from user input.
func GenPrivKeySecp256k1FromSecret(secret []byte) PrivKeySecp256k1 {
	privKey32 := Sha256(secret) // Not Ripemd160 because we want 32 bytes.
	priv, _ := secp256k1.PrivKeyFromBytes(secp256k1.S256(), privKey32)
	privKeyBytes := [32]byte{}
	copy(privKeyBytes[:], priv.Serialize())
	return PrivKeySecp256k1(privKeyBytes)
}
