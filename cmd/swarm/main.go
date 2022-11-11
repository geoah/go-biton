package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ghodss/yaml"
	"github.com/kelseyhightower/envconfig"

	"github.com/geoah/go-biton"
)

func main() {
	nodeInfo := biton.NodeInfo{}

	err := envconfig.Process("BITON", &nodeInfo)
	if err != nil {
		log.Fatalf("failed to process envconfig: %v", err)
	}

	if nodeInfo.KeyPair.Public == nil {
		nodeInfo.KeyPair, err = biton.NewKeypair()
		if err != nil {
			log.Fatalf("failed to create keypair: %v", err)
		}
	}

	mainlineDHT, err := biton.NewMainlineDHT(nodeInfo)
	if err != nil {
		log.Fatalf("failed to create mainline DHT: %v", err)
	}

	yamlBytes, _ := yaml.Marshal(nodeInfo)
	fmt.Println("node info:")
	fmt.Println(string(yamlBytes))

	fmt.Println("* mainline node id:", biton.NewMainlineNodeID(nodeInfo.KeyPair.Public))

	swarm, err := biton.NewSwarm(
		nodeInfo,
		mainlineDHT,
		biton.SwarmGlobalSeed,
		biton.SwarmGlobalPath,
	)
	if err != nil {
		log.Fatalf("failed to create swarm: %v", err)
	}

	err = swarm.Listen(context.Background())
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	err = swarm.Bootstrap(context.Background())
	if err != nil {
		log.Fatalf("failed to bootstrap: %v", err)
	}

	fmt.Println("Found peers:")
	for _, peer := range swarm.ListPeers() {
		fmt.Printf("* %x: %v\n", peer.ID, peer.Addresses)
	}

	fmt.Println("Press enter to exit...")
	fmt.Scanln()
}
