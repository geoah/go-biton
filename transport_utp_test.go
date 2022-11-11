package biton

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransportUTP_E2E(t *testing.T) {
	ctx := context.Background()

	transport := TransportUTP{}
	listener, err := transport.Listen("127.0.0.1:0")
	require.NoError(t, err)

	go func() {
		conn, err := listener.Accept()
		require.NoError(t, err)
		defer conn.Close()

		_, err = conn.Write([]byte("hello"))
		require.NoError(t, err)
	}()

	conn, err := transport.Dial(ctx, listener.Addr().String())
	require.NoError(t, err)
	defer conn.Close()

	buf := make([]byte, 5)
	n, err := conn.Read(buf)
	require.NoError(t, err)
	require.Equal(t, 5, n)
	require.Equal(t, "hello", string(buf))
}
