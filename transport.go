package biton

import (
	"context"
	"net"
)

type Transport interface {
	Dial(ctx context.Context, address string) (net.Conn, error)
	Listen(address string) (net.Listener, error)
	Network() string
}
