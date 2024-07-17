package main

import (
	"context"
	"fmt"
	"github.com/czh0526/libp2p-examples/utils"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var idht *dht.IpfsDHT
	priv, err := utils.GeneratePrivateKey("privkey.pem")

	h, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/9000",
			"/ip4/0.0.0.0/udp/9000/quic"),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			idht, err = dht.New(context.Background(), h)
			return idht, err
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
	addrs, err := peer.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		panic(err)
	}
	fmt.Printf("libp2p peer.ID = %v, peerAddrs = %v, node address: %v \n",
		h.ID(), h.Addrs(), addrs)

	<-ctx.Done()

}
