package main

import (
	"context"
	"fmt"
	"github.com/czh0526/libp2p-examples/utils"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rhost, err := makeHost()
	if err != nil {
		log.Printf("Failed to create host: %v", err)
	}

	_, err = relay.New(rhost)
	if err != nil {
		log.Printf("Failed to instantiate the relay1: %v", err)
		return
	}

	fmt.Printf("peer.ID = %v\n", rhost.ID())
	fmt.Println("peer addresses: ")
	for _, addr := range rhost.Addrs() {
		fmt.Printf("\t=> %v\n", addr)
	}

	<-ctx.Done()
}

func makeHost() (host.Host, error) {
	priv, err := utils.GeneratePrivateKey("privkey.pem")
	if err != nil {
		panic(fmt.Sprintf("get private key failed: err = %v", err))
	}

	host, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/8000"),
		libp2p.EnableRelay(),
	)
	if err != nil {
		log.Printf("Failed to create relay1, err = %v", err)
	}

	return host, err
}
