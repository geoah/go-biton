package biton

import (
	"context"
	"crypto/ed25519"
	"fmt"

	"github.com/anacrolix/dht/v2"
	"github.com/anacrolix/dht/v2/krpc"
)

const (
	webtorrentPeerIDPrefix = "-WW0102-"
)

type MainlineDHT interface {
	GetPeers(ctx context.Context, transport, swarmAddress string) (<-chan PeerAddress, error)
}

type mainlineDHT struct {
	nodeInfo NodeInfo
	dhtNode  *dht.Server
}

func NewMainlineNodeID(pub ed25519.PublicKey) krpc.ID {
	var nodeID krpc.ID
	copy(nodeID[:], webtorrentPeerIDPrefix)
	copy(nodeID[len(webtorrentPeerIDPrefix):], pub[:])
	return nodeID
}

func NewMainlineDHT(nodeInfo NodeInfo) (MainlineDHT, error) {
	// construct new dht config
	cfg := dht.NewDefaultServerConfig()

	// update the config with the peer id
	cfg.NodeId = NewMainlineNodeID(nodeInfo.KeyPair.Public)

	// construct a new dht server
	dhtNode, err := dht.NewServer(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create dht server: %w", err)
	}

	fmt.Println("* mainline address:", dhtNode.Addr().String())

	r := &mainlineDHT{
		nodeInfo: nodeInfo,
		dhtNode:  dhtNode,
	}
	return r, nil
}

func (r *mainlineDHT) GetPeers(ctx context.Context, transport, swarmAddress string) (<-chan PeerAddress, error) {
	// construct infohash from swarm address and transport
	ih := NewInfoHash(fmt.Sprintf("%s/%s", transport, swarmAddress))

	// TODO: debugging, remove
	fmt.Printf("* swarm infohash: %x\n", ih)

	// announce to the dht, and get back peers
	a, err := r.dhtNode.AnnounceTraversal(
		ih,
		dht.AnnouncePeer(
			dht.AnnouncePeerOpts{
				ImpliedPort: false,
				// TODO: we should be announceing a different port based on the transport
				Port: r.nodeInfo.UTPPort,
			},
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to announce traversal: %w", err)
	}

	// start a channel to send found addresses to
	addrs := make(chan PeerAddress, 1)

	// keep track of the addresses we've seen
	seen := map[string]struct{}{}

	// Note from BEP 0005:
	//
	// If the queried node has peers for the infohash, they are returned in a
	// key "values" as a list of strings.
	//
	// If the queried node has no peers for the infohash, a key "nodes" is
	// returned containing the K nodes in the queried nodes routing table
	// closest to the infohash supplied in the query.

	go func() {
		defer func() {
			// wait for the traversal to finish
			<-a.Finished()

			// close the channel
			close(addrs)

			// stop announcing
			// TODO: decouple from get peers
			a.Close()
		}()

		for {
			select {
			case <-ctx.Done():
				a.StopTraversing()
				return
			case ps, ok := <-a.Peers:
				if !ok {
					return
				}

				// Note: ps.Values contains peers that should have the infohash
				// Note: ps.Nodes contains peers that are close to the infohash

				// TODO: debugging, remove
				// fmt.Println("ps.Values:", ps.Values)
				// fmt.Println("ps.Nodes:", ps.Nodes)

				// if there are no peers of interest, move on
				if len(ps.Values) == 0 {
					continue
				}

				// convert the addresses to PeerAddresses
				for _, p := range ps.Values {
					// if we've already seen this address, move on
					if _, ok := seen[p.String()]; ok {
						continue
					}

					// ignore port 1
					if p.Port == 1 {
						continue
					}

					// else push the address to the channel
					addrs <- PeerAddress{
						transport: transport,
						address:   p.String(),
					}
					// and mark it as seen
					seen[p.String()] = struct{}{}
				}
			}
		}
	}()

	return addrs, nil
}
