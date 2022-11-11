package biton

import (
	"net"
)

type (
	PeerID         string
	PeerMainlineID [20]byte
)

type (
	Peer struct {
		Info PeerInfo
		Conn net.Conn
	}
	PeerInfo struct {
		ID        PeerID
		PublicKey []byte
		Addresses []PeerAddress
	}
	PeerAddress struct {
		address   string
		transport string
	}
)

func MergePeerInfo(a, b PeerInfo) PeerInfo {
	c := PeerInfo{
		ID:        a.ID,
		PublicKey: a.PublicKey,
		Addresses: a.Addresses,
	}
	for _, addr := range b.Addresses {
		if !c.HasAddress(addr) {
			c.Addresses = append(c.Addresses, addr)
		}
	}
	return c
}

func (a *PeerAddress) Network() string {
	return a.transport
}

func (a *PeerAddress) String() string {
	return a.address
}

func (p PeerInfo) HasAddress(address PeerAddress) bool {
	for _, a := range p.Addresses {
		if a == address {
			return true
		}
	}
	return false
}

func NewPeerAddress(transport, address string) PeerAddress {
	return PeerAddress{
		transport: transport,
		address:   address,
	}
}
