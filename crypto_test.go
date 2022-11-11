package biton

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeyPair_Base58(t *testing.T) {
	// create new keypair
	pair, err := NewKeypair()
	require.NoError(t, err)

	// marshal to base58
	base58, err := pair.MarshalText()
	require.NoError(t, err)

	// unmarshal from base58
	var pair2 KeyPair
	err = pair2.UnmarshalText(base58)
	require.NoError(t, err)

	// check that the two match
	require.Equal(t, pair.Private, pair2.Private)
	require.Equal(t, pair.Public, pair2.Public)
}
