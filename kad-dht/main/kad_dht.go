package main

import (
	"context"
	"fmt"
	"github.com/czh0526/libp2p-examples/utils"
	"github.com/libp2p/go-libp2p"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/routing"
)

func createNode() (host.Host, error) {

	priv, err := utils.GeneratePrivateKey("privkey.pem")
	if err != nil {
		return nil, err
	}
	host, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/8080"),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			dht, err := kaddht.New(context.Background(), h, kaddht.Mode(kaddht.ModeServer))
			return dht, err
		}),
	)
	if err != nil {
		return nil, err
	}

	fmt.Printf("libp2p peer.ID = %v\n", host.ID())
	fmt.Println("libp2p peer addresses: ")
	for _, addr := range host.Addrs() {
		fmt.Printf("\t=> %v\n", addr)
	}

	return host, nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	host, err := createNode()
	if err != nil {
		panic(fmt.Sprintf("create node 1 failed: %v", err))
	}
	defer host.Close()

	<-ctx.Done()
}
