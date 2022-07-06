# try-libp2p-tor

Small program that allows you to run a libp2p node both with and without the tor transport, and connect a node to another node.

## Build

This program relies on my fork of the go-libp2p-tor-transport.

Clone it into your GOPATH and ensure the `replace` directive in `go.mod` points to the local fork:

```
git clone https://github.com/noot/go-libp2p-tor-transport
```

Then build:
```
go build -tags=embedTor 
```

## Usage

To run a node with the tor transport:
```bash
./try-libp2p-tor
# []
```

To run a default libp2p node (without the tor transport):
```bash
./try-libp2p-tor --no-tor
# [/ip4/192.168.0.102/tcp/62049/p2p/QmPYWg7LX1r4bBPgSb4u2cPRMzmXCuKFEa32KYGLmzP2yU /ip4/127.0.0.1/tcp/62049/p2p/QmPYWg7LX1r4bBPgSb4u2cPRMzmXCuKFEa32KYGLmzP2yU]
```

To connect a node to other nodes:
```bash
./try-libp2p-tor --bootnodes /ip4/127.0.0.1/tcp/62049/p2p/QmPYWg7LX1r4bBPgSb4u2cPRMzmXCuKFEa32KYGLmzP2yU
```
