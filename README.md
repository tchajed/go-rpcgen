# RFC4506 XDR generator for Go

Run `make` to build the `go-rpcgen` tool and compile several specs,
including NFS (rfc1813) and SUNRPC (rfc1057).

There is an example client in `example/main.go` that connects to an
NFS server and issues some NFS RPCs.

## TODO

- Validate that encoded/decoded enums match one of the allowed values.
