package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	libp2pvpn "github.com/balena/go-libp2p-vpn"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

func main() {
	keygenCmd := flag.NewFlagSet("keygen", flag.ExitOnError)
	keygenFileF := keygenCmd.String("f", "", "Specifies the filename of the private key file")
	keygenKeyTypeF := keygenCmd.String("t", "ecdsa", "Specifies the type of key to create. "+
		"The possible values are \"ecdsa\", \"ed25519\", \"secp256k1\", or \"rsa\"")

	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	runPrivKeyF := runCmd.String("f", "", "Specifies the filename of the private key file")
	runPeerF := runCmd.String("p", "", "Specifies the peer address")
	runListenAddrsF := runCmd.String("l", "", "Specifies listen addresses")
	runTunnelAddrsF := runCmd.String("t", "", "Specifies the tunnel addresses")

	switch os.Args[1] {
	case "keygen":
		keygenCmd.Parse(os.Args[2:])
		runKeyGen(*keygenKeyTypeF, *keygenFileF)
	case "run":
		runCmd.Parse(os.Args[2:])
		tunnelAddrs := strings.Split(*runTunnelAddrsF, ",")
		listenAddrs := strings.Split(*runListenAddrsF, ",")
		run(*runPrivKeyF, *runPeerF, tunnelAddrs[0], tunnelAddrs[1], listenAddrs)
	}
}

func runKeyGen(keyType, file string) {
	var typ int

	switch keyType {
	case "ecdsa":
		typ = crypto.ECDSA
	case "ed25519":
		typ = crypto.Ed25519
	case "secp256k1":
		typ = crypto.Secp256k1
	default:
		panic(fmt.Sprintf("Unknown key type '%s'", keyType))
	}

	privKey, pubKey, err := crypto.GenerateKeyPair(typ, 0)
	if err != nil {
		panic(err)
	}

	id, err := peer.IDFromPublicKey(pubKey)
	if err != nil {
		panic(err)
	}

	b, err := crypto.MarshalPrivateKey(privKey)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(file, b, 0600)
	if err != nil {
		panic(err)
	}

	fmt.Printf("peer ID: %s\n", id.Pretty())
}

func run(privKeyFile, peerStr, localAddr, remoteAddr string, listenAddrs []string) {
	ctx := context.Background()

	b, err := os.ReadFile(privKeyFile)
	if err != nil {
		panic(err)
	}

	privKey, err := crypto.UnmarshalPrivateKey(b)
	if err != nil {
		panic(err)
	}

	h, err := libp2p.New(
		libp2p.ListenAddrStrings(listenAddrs...),
		libp2p.Identity(privKey),
	)
	if err != nil {
		panic(err)
	}

	peerAddr, err := ma.NewMultiaddr(peerStr)
	if err != nil {
		panic(err)
	}

	peerAddrInfo, err := peer.AddrInfoFromP2pAddr(peerAddr)
	if err != nil {
		panic(err)
	}

	vpn, err := libp2pvpn.New(
		peerAddrInfo.ID,
		libp2pvpn.TunnelIP(localAddr, remoteAddr),
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Trying to connect to %s\n", peerStr)
	for {
		err = h.Connect(ctx, *peerAddrInfo)
		if err == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}
	fmt.Println("[x] Connection succeeded!")

	h.SetStreamHandler(vpn.Protocol(), vpn.Handler())

	vpn.Serve(ctx, h)
}
