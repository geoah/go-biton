package biton

import (
	"encoding/hex"

	"github.com/mr-tron/base58/base58"
	"golang.org/x/crypto/blake2b"
)

type InfoHash [20]byte

func (ih InfoHash) HexString() string {
	return hex.EncodeToString(ih[:])
}

func (ih InfoHash) Base58String() string {
	return base58.Encode(ih[:])
}

func (ih InfoHash) String() string {
	return ih.HexString()
}

func NewInfoHash(s string) InfoHash {
	b := blake2b.Sum256([]byte(s))
	var ih InfoHash
	copy(ih[:], b[0:20])
	return ih
}
