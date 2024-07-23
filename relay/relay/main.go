package main

import (
	"context"
	"fmt"
	"github.com/czh0526/libp2p-examples/utils"
	"github.com/libp2p/go-libp2p"
	"log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	priv, err := utils.GeneratePrivateKey("privkey.pem")
	if err != nil {
		panic(fmt.Sprintf("get private key failed: err = %v", err))
	}

	host, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/8080"),
		libp2p.EnableRelay(),
	)
	if err != nil {
		log.Printf("Failed to create relay1, err = %v", err)
	}

	//_, err = relay.New(host)
	//if err != nil {
	//	log.Printf("Failed to instantiate the relay1: %v", err)
	//	return
	//}

	fmt.Printf("peer.ID = %v\n", host.ID())
	fmt.Println("peer addresses: ")
	for _, addr := range host.Addrs() {
		fmt.Printf("\t=> %v\n", addr)
	}

	<-ctx.Done()
}
