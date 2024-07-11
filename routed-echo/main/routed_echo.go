package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	ds "github.com/ipfs/go-datastore"
	dsync "github.com/ipfs/go-datastore/sync"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	rhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	ma "github.com/multiformats/go-multiaddr"
	"io"
	mrand "math/rand"
)

func main() {
	listenF := flag.Int("l", 0, "listen port")
	target := flag.String("d", "", "target peer to dial")
	seed := flag.Int64("seed", 0, "set random seed for id generation")
	global := flag.Bool("global", false, "use global ipfs peers for bootstrapping")
	flag.Parse()

	if *listenF == 0 {
		panic("Please provide a port to bind on with -l")
	}

	var bootstrapPeers []peer.AddrInfo
	var globalFlag string
	if *global {
		fmt.Println("using global bootstrap")
		bootstrapPeers = IPFS_PEERS
		globalFlag = " -global"
	} else {
		fmt.Println("using local bootstrap")
		bootstrapPeers = getLocalPeerInfo()
		globalFlag = ""
	}
	ha, err := makeRoutedHost(*listenF, *seed, bootstrapPeers, globalFlag)
	if err != nil {
		panic(fmt.Sprintf("make routed host failed: err = %v", err))
	}

	ha.SetStreamHandler("/echo/1.0.0", func(s network.Stream) {
		fmt.Printf("new echo stream from %s\n", s.Conn().RemotePeer())
		if err := doEcho(s); err != nil {
			fmt.Printf("echo failed: err = %v\n", err)
			s.Reset()
		} else {
			s.Close()
		}
	})

	if *target == "" {
		fmt.Println("Listening for connections")
		select {}
	} else {
		peerid, err := peer.Decode(*target)
		if err != nil {
			panic(fmt.Sprintf("peer(`%v`) decode failed: err = %s", *target, err))
		}

		fmt.Println("opening stream:")
		s, err := ha.NewStream(context.Background(), peerid, protocol.ID("/echo/1.0.0"))
		if err != nil {
			panic(fmt.Sprintf("peer(`%v`) create stream failed: err = %v", *target, err))
		}

		_, err = s.Write([]byte("hello world!\n"))
		if err != nil {
			panic(fmt.Sprintf("peer(`%v`) write stream failed: err = %v", *target, err))
		}

		out, err := io.ReadAll(s)
		if err != nil {
			panic(fmt.Sprintf("peer(`%v`) read stream failed: err = %v", *target, err))
		}

		fmt.Printf("read reply: %q\n", out)
	}
}

func makeRoutedHost(listenPort int, randSeed int64, bootstrapPeers []peer.AddrInfo,
	globalFlag string) (host.Host, error) {

	var r io.Reader
	if randSeed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(randSeed))
	}

	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.ECDSA, 2048, r)
	if err != nil {
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenPort)),
		libp2p.Identity(priv),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.NATPortMap(),
	}

	ctx := context.Background()
	basicHost, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}

	// make the routed host
	dstore := dsync.MutexWrap(ds.NewMapDatastore())
	dht := dht.NewDHT(ctx, basicHost, dstore)
	routedHost := rhost.Wrap(basicHost, dht)

	// connect to the ipfs nodes
	err = bootstrapConnect(ctx, routedHost, bootstrapPeers)
	if err != nil {
		return nil, err
	}

	// bootstrap the host
	err = dht.Bootstrap(ctx)
	if err != nil {
		return nil, err
	}

	// build host multiaddr
	hostAddr, err := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", routedHost.ID()))
	if err != nil {
		return nil, err
	}

	addrs := routedHost.Addrs()
	fmt.Println("I can be reached at:")
	for _, addr := range addrs {
		fmt.Printf("\t%s\n", addr.Encapsulate(hostAddr))
	}
	fmt.Printf("Now run `./routed-echo -l %d %s%s` on a different terminal\n",
		listenPort+1, routedHost.ID(), globalFlag)

	return routedHost, nil
}

func doEcho(s network.Stream) error {
	return nil
}
