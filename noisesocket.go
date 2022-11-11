package biton

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
)

// noisesocket does something weird and marshals the peer public key
// into a JSON string, and returns it in the conn.ConnectionState()
// this helper function extracts the peer public key from it
func noisePeerPublicKeyFromConnectionState(state tls.ConnectionState) (PublicKey, error) {
	body := struct {
		PeerPublic    []byte
		HandshakeHash []byte
	}{}

	err := json.Unmarshal(state.TLSUnique, &body)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall tls unique: %w", err)
	}

	return body.PeerPublic, nil
}
