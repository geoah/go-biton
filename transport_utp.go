package biton

import (
	"context"
	"fmt"
	"net"

	"github.com/neilalexander/utp"
)

type TransportUTP struct{}

func (t *TransportUTP) Dial(ctx context.Context, address string) (net.Conn, error) {
	c, err := utp.DialContext(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}
	return c, nil
}

func (t *TransportUTP) Listen(address string) (net.Listener, error) {
	l, err := utp.Listen(address)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}
	return l, nil
}

func (t *TransportUTP) Network() string {
	return "utp"
}
