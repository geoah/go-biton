# go-biton

Proof of concept implementation of the [biton] protocol.
Please do not use as this is still under heavy development.

## Running swarm demo peer

Available env vars:

* `KEYPAIR` x25519 KeyPair in base58 format, will generate new one if empty.
* `UTP_HOST` UTP host to advertise on mainline, defaults to `0.0.0.0`.
* `UTP_PORT` UTP port to advertise on mainline, defaults to `0`.
* `MAINLINE_HOST` UDP port to use for mainline, defaults to `0.0.0.0`.
* `MAINLINE_PORT` UDP port to use for mainline, defaults to `6881`.

```sh
BITON_UTP_PORT=9999 go run cmd/swarm/main.go
```

## Notes

* Currently using noise's XX pattern. Not sure if that's the right one to use.
* Had to fork `noisesocket` to allow wrapping `net.Conn` and `net.Listener`.
  The `go.mod` replace should be removed in order to allow importing the lib.

---

[biton]: https://bitonproject.org
