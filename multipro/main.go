package main

import (
	"context"
	"fmt"
	"github.com/czh0526/libp2p-examples/utils"
	"github.com/libp2p/go-libp2p"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/peerstore"
	rhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	ma "github.com/multiformats/go-multiaddr"
	"math/rand"
)

func main() {
	rnd := rand.New(rand.NewSource(100))
	port1 := rnd.Intn(100) + 10000
	port2 := port1 + 1

	done := make(chan bool, 1)
	h1 := makeNode(port1, done)
	h2 := makeNode(port2, done)

	run(h1, h2, done)
}

func makeNode(port int, done chan bool) *Node {
	ctx := context.Background()

	// 读取固定的私钥文件
	priv, err := utils.GeneratePrivateKey("privkey.pem")
	if err != nil {
		panic(err)
	}

	// 构建 BasicHost
	listen, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
	basicHost, _ := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrs(listen),
	)

	// 构建 DHT
	dht, err := kaddht.New(ctx, basicHost)
	if err != nil {
		panic(fmt.Sprintf("new dht failed: err = %v", err))
	}
	routedHost := rhost.Wrap(basicHost, dht)

	// bootstrap the host
	err = dht.Bootstrap(ctx)
	if err != nil {
		panic(fmt.Sprintf("host bootstrap failed, err = %v", err))
	}

	// connect to the ipfs nodes
	err = bootstrapConnect(ctx, routedHost, BOOTSTRAP_PEERS)
	if err != nil {
		panic(fmt.Sprintf("connect bootstrap peers failed, err = %v", err))
	}

	return NewNode(routedHost, done)
}

func run(h1, h2 *Node, done <-chan bool) {
	h1.Peerstore().AddAddrs(h2.ID(), h2.Addrs(), peerstore.PermanentAddrTTL)
	h2.Peerstore().AddAddrs(h1.ID(), h1.Addrs(), peerstore.PermanentAddrTTL)

	h1.Ping(h2.Host)
	h2.Ping(h1.Host)
	//h1.Echo(h2.Host)
	//h2.Echo(h1.Host)

	for i := 0; i < 4; i++ {
		<-done
	}
}
