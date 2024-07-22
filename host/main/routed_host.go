package main

import (
	"context"
	"fmt"
	"github.com/czh0526/libp2p-examples/utils"
	"github.com/libp2p/go-libp2p"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var dht *kaddht.IpfsDHT
	priv, err := utils.GeneratePrivateKey("privkey.pem")

	h, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/4001",
			"/ip4/0.0.0.0/udp/4001/quic-v1"),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			dht, err = kaddht.New(context.Background(), h, kaddht.Mode(kaddht.ModeServer))
			return dht, err
		}),
	)
	if err != nil {
		panic(err)
	}
	defer h.Close()

	peerInfo := peer.AddrInfo{
		ID:    h.ID(),
		Addrs: h.Addrs(),
	}
	maddrs, err := peer.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		panic(err)
	}
	fmt.Printf("libp2p peer.ID = %v\n", h.ID())
	fmt.Printf("libp2p peer.Addrs = %v\n", h.Addrs())
	fmt.Printf("libp2p multi address = %v\n", maddrs)

	<-ctx.Done()

}
