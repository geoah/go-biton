package biton

import (
	"crypto/rand"
	"encoding/json"
	"fmt"

	"github.com/flynn/noise"
	"github.com/mr-tron/base58/base58"
)

type (
	KeyPair struct {
		Public  []byte
		Private []byte
	}
	PublicKey  []byte
	PrivateKey []byte
)

func NewKeypair() (KeyPair, error) {
	pair, err := noise.DH25519.GenerateKeypair(rand.Reader)
	if err != nil {
		return KeyPair{}, fmt.Errorf("failed to generate keypair: %w", err)
	}
	return KeyPair{
		Public:  pair.Public,
		Private: pair.Private,
	}, nil
}

func NewKeypairFromBase58(keyPairBase string) (KeyPair, error) {
	keyPair := KeyPair{}
	err := keyPair.UnmarshalText([]byte(keyPairBase))
	return keyPair, err
}

func (kp KeyPair) MarshalJSON() ([]byte, error) {
	r, err := kp.MarshalText()
	if err != nil {
		return nil, err
	}
	return json.Marshal(string(r))
}

func (kp *KeyPair) MarshalText() ([]byte, error) {
	return []byte(base58.Encode(
		append(
			kp.Public,
			kp.Private...,
		),
	)), nil
}

func (kp *KeyPair) UnmarshalText(keyPairBase []byte) error {
	keyPairBytes, err := base58.Decode(string(keyPairBase))
	if err != nil {
		return fmt.Errorf("failed to decode base58: %w", err)
	}

	kp.Public = keyPairBytes[:len(keyPairBytes)/2]
	kp.Private = keyPairBytes[len(keyPairBytes)/2:]

	return nil
}

func (kp *KeyPair) DHKey() noise.DHKey {
	return noise.DHKey{
		Public:  kp.Public,
		Private: kp.Private,
	}
}

func (p PublicKey) Identity() PeerID {
	return PeerID(base58.Encode(p))
}
