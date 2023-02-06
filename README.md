# go-libp2p-vpn

> go-libp2p's VPN

Package `go-libp2p-vpn` is a VPN over libp2p.

## Install

```sh
go get github.com/balena/go-libp2p-vpn
```

## How to use

```Go
package main

import (
  "github.com/libp2p/go-libp2p"
  "github.com/balena/go-libp2p-vpn"
)

func main() {
  // Create the libp2p host
  host, err := libp2p.New()

  // Then create the VPN interface
  vpn, _ := libp2pvpn.New(peer)
 
  // Now set the stream handler (p2p->if)
  host.SetStreamHandler(vpn.Protocol(), vpn.Handler())

  // And serve packets (if->p2p)
  vpn.Serve(ctx, host)
}
```

Above example is overly simplified, double check the options passed to `libp2p.New()` and to `libp2pvpn.New()`.

## Contribute

Feel free to join in. All welcome. Open an [issue](https://github.com/balena/go-libp2p-vpn/issues)!

This repository falls under the libp2p [Code of Conduct](https://github.com/libp2p/community/blob/master/code-of-conduct.md).

### Want to hack on libp2p?

[![](https://cdn.rawgit.com/libp2p/community/master/img/contribute.gif)](https://github.com/libp2p/community/blob/master/CONTRIBUTE.md)

## License

[MIT](LICENSE) © 2023 Guilherme Versiani
