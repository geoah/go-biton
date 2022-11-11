package biton

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/flynn/noise"
	"gopkg.in/noisesocket.v0"
)

const (
	Version = "biton0"
)

var (
	NetworkMagicMain string = ""
	NetworkMagicTest string = "test"
)

const (
	SwarmGlobalSeed string = ""
	SwarmGlobalPath string = ""
)

type Swarm struct {
	lock          sync.RWMutex
	netMagic      string
	swarmSeed     string
	swarmPath     string
	swarmPeers    map[PeerID]Peer
	transports    swarmTransports
	mainlineDHT   MainlineDHT
	mainlineAddrs chan PeerAddress
	nodeInfo      NodeInfo
	noiseDHKey    noise.DHKey
}

// TODO: until we have multiple transports, this should enough
type swarmTransports struct {
	utp Transport
}

func NewSwarm(
	nodeInfo NodeInfo,
	mainlineDHT MainlineDHT,
	swarmSeed string,
	swarmPath string,
) (*Swarm, error) {
	s := &Swarm{
		netMagic:   NetworkMagicMain,
		swarmSeed:  swarmSeed,
		swarmPath:  swarmPath,
		swarmPeers: map[PeerID]Peer{},
		transports: swarmTransports{
			utp: &TransportUTP{},
		},
		mainlineDHT:   mainlineDHT,
		mainlineAddrs: make(chan PeerAddress, 1),
		nodeInfo:      nodeInfo,
		noiseDHKey: noise.DHKey{
			Private: nodeInfo.KeyPair.Private,
			Public:  nodeInfo.KeyPair.Public,
		},
	}
	go s.handleMainlineAddrs()
	return s, nil
}

func (s *Swarm) Address() string {
	return fmt.Sprintf(
		"%s:%s:%s:%s",
		Version,
		s.netMagic,
		s.swarmPath,
		s.swarmSeed,
	)
}

// Bootstrap connects to the mainline DHT and starts looking for peers
func (s *Swarm) Bootstrap(ctx context.Context) error {
	transports := []string{
		s.transports.utp.Network(),
	}
	// TODO: once we have multiple transports, we should do this in parallel
	for _, transport := range transports {
		addrs, err := s.mainlineDHT.GetPeers(
			ctx,
			transport,
			s.Address(),
		)
		if err != nil {
			return fmt.Errorf("failed to get peers: %w", err)
		}
		for addr := range addrs {
			s.mainlineAddrs <- addr
		}
	}

	return nil
}

func (s *Swarm) ListPeers() []PeerInfo {
	s.lock.RLock()
	defer s.lock.RUnlock()

	peers := []PeerInfo{}
	for _, p := range s.swarmPeers {
		peers = append(peers, p.Info)
	}
	return peers
}

func (s *Swarm) handleMainlineAddrs() {
	// for each address we get from the dht
	for addr := range s.mainlineAddrs {
		// TODO: debugging, remove
		fmt.Printf("* got peer address: %s\n", addr.String())

		// check if a peer with this address already exists
		// TODO: keep an async map of connected addrs instead
		found := false
		s.lock.RLock()
		for _, p := range s.swarmPeers {
			if p.Info.HasAddress(addr) {
				found = true
				break
			}
		}
		s.lock.RUnlock()

		// if found, move on
		if found {
			continue
		}

		// else dial it
		// TODO: dial stores the peer, but should it?
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_, err := s.dial(ctx, addr)
		if err != nil {
			log.Println("failed to dial peer", addr, err)
			cancel()
			continue
		}
		cancel()
	}
}

func (s *Swarm) Listen(ctx context.Context) error {
	// start utp listener
	utpAddr := fmt.Sprintf("%s:%d", s.nodeInfo.UTPHost, s.nodeInfo.UTPPort)
	utpLst, err := s.transports.utp.Listen(utpAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on utp: %w", err)
	}

	// start listening for connections
	go s.handleListener(ctx, utpLst)

	return nil
}

func (s *Swarm) handleListener(
	ctx context.Context,
	lstRaw net.Listener,
) {
	// wrap the listener in a noise listener
	lst := noisesocket.WrapListener(
		&noisesocket.ConnectionConfig{
			StaticKey: s.nodeInfo.KeyPair.DHKey(),
		},
		lstRaw,
	)

	// start accepting connections
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := lst.Accept()
			if err != nil {
				log.Printf("failed to accept: %v", err)
				continue
			}
			// TODO: debugging, remove
			fmt.Printf("* accepted connection from addr %s\n", conn.RemoteAddr())
			go s.handleConnection(conn.(*noisesocket.Conn))
		}
	}
}

func (s *Swarm) handleConnection(conn *noisesocket.Conn) (Peer, error) {
	// get the peer's public key
	peerPubKey, err := noisePeerPublicKeyFromConnectionState(
		conn.ConnectionState(),
	)
	if err != nil {
		return Peer{}, fmt.Errorf("failed to get peer public key: %w", err)
	}

	// construct the peer
	peer := Peer{
		Info: PeerInfo{
			ID:        peerPubKey.Identity(),
			PublicKey: peerPubKey,
			Addresses: []PeerAddress{},
		},
		Conn: conn,
	}

	// TODO: debugging, remove
	// wait for ping message from client
	body := make([]byte, 4)
	_, err = conn.Read(body)
	if err != nil {
		return Peer{}, fmt.Errorf("failed to read ping message: %w", err)
	}

	fmt.Println("🔥 got ping message")

	// store the peer
	s.storePeer(peer)

	// TODO: debugging, remove
	fmt.Printf("* connected (in) to peer: %s\n", peer.Info.ID)

	return peer, nil
}

func (s *Swarm) dial(ctx context.Context, addr PeerAddress) (Peer, error) {
	peer := Peer{}

	// TODO: find a way to skip dialing our own peer

	// TODO: debugging, remove
	// hairpin allows us to dial other peers on the same machine, by update
	// the found address to our local one
	host, port, _ := net.SplitHostPort(addr.String())
	if host == s.nodeInfo.HairpinHost {
		// at the same time, skip dialing if it's our own port
		if port == fmt.Sprintf("%d", s.nodeInfo.UTPPort) {
			return peer, fmt.Errorf("hairpin: skipping self")
		}
		fmt.Println("DEBUG: found hairpin host:", host)
		addr.address = fmt.Sprintf("127.0.0.1:%s", port)
	}

	// get transport
	var transport Transport
	switch addr.Network() {
	case "utp":
		transport = s.transports.utp
	default:
		return peer, fmt.Errorf("unknown transport: %s", addr.Network())
	}

	// TODO: debugging, remove
	fmt.Printf("* dialling addr %s\n", addr.String())

	// dial the peer using the transport above
	connRaw, err := transport.Dial(ctx, addr.String())
	if err != nil {
		return peer, fmt.Errorf("failed to dial: %w", err)
	}

	// wrap the raw connection in a noise socket
	conn := noisesocket.WrapConn(
		&noisesocket.ConnectionConfig{
			StaticKey: s.noiseDHKey,
		},
		connRaw,
		true,
	)

	// TODO: consider whether we want to handshake now or later
	// handshake is normally called when sending the first message,
	// instead we call it here to make sure the connection is ready
	// err = conn.Handshake()
	// if err != nil {
	// 	return peer, fmt.Errorf("failed to handshake: %w", err)
	// }

	// get the peer's public key
	peerPubKey, err := noisePeerPublicKeyFromConnectionState(
		conn.ConnectionState(),
	)
	if err != nil {
		return peer, fmt.Errorf("failed to get peer public key: %w", err)
	}

	// construct the peer
	peer = Peer{
		Info: PeerInfo{
			ID:        peerPubKey.Identity(),
			PublicKey: peerPubKey,
			Addresses: []PeerAddress{
				addr,
			},
		},
		Conn: conn,
	}

	// TODO: debugging, remove
	// ping the peer
	_, err = conn.Write([]byte("ping"))
	if err != nil {
		return peer, fmt.Errorf("failed to ping peer: %w", err)
	}

	fmt.Println("🔥 sent ping message")

	// store the peer
	s.storePeer(peer)

	// TODO: debugging, remove
	fmt.Printf("* connected (out) to peer: %s\n", peer.Info.ID)

	return peer, nil
}

func (s *Swarm) storePeer(peer Peer) {
	// TODO: first check if a peer with this ID or address already exists
	s.lock.Lock()
	s.swarmPeers[peer.Info.ID] = peer
	s.lock.Unlock()
}
