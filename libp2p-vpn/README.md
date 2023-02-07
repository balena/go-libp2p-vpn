# Simple VPN with go-libp2p-vpn

This is an example that quickly shows how to use the `go-libp2p-vpn` stack.

The app will start a local TUN interface and wait for a connection to the peer on the `/vpn/tun/1.0.0` protocol.

After connecting, the VPN is fully established.

## Build

From the `libp2p-vpn` directory run the following:

```
> cd libp2p-vpn
> go build
```

## Usage

First create two private keys, it will print the respective peer IDs:

```
$ ./libp2p-vpn keygen -f peer1
peer ID: Qm..Ana
$ ./libp2p-vpn keygen -f peer2
peer ID: Qm..Bob
```

On Mac, in a terminal do:

```
$ sudo ./libp2p-vpn run \
    -f peer1 \
    -l /ip4/127.0.0.1/tcp/10000 \
    -p /ip4/127.0.0.1/tcp/10001/p2p/Qm..Bob \
    -t 10.0.0.1,10.0.0.2
```

then from another terminal:

```
$ sudo ./libp2p-vpn run \
    -f peer2 \
    -l /ip4/127.0.0.1/tcp/10001 \
    -p /ip4/127.0.0.1/tcp/10000/p2p/Qm..Ana \
    -t 10.0.0.2,10.0.0.1
```
