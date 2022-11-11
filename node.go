package biton

type (
	NodeInfo struct {
		// KeyPair for noise
		KeyPair KeyPair `envconfig:"KEYPAIR"`

		// Swarm UTP configuration
		UTPHost string `envconfig:"UTP_HOST" default:"0.0.0.0"`
		UTPPort int    `envconfig:"UTP_PORT" default:"0"`

		// Mainline DHT configuration
		MainlineHost string `envconfig:"MAINLINE_HOST" default:"0.0.0.0"`
		MainlinePort int    `envconfig:"MAINLINE_PORT" default:"6881"`

		// TODO: debugging, remove
		// HairpinHost is used for local debugging
		// When the mainline DHT sees this address it will replace it with
		// a local address to allow peers on the same machine to connect to
		// each other
		HairpinHost string `envconfig:"DEBUG_HAIRPIN_HOST"`
	}
)
