package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/czh0526/libp2p-examples/utils"
	"github.com/libp2p/go-libp2p"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	rhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	ma "github.com/multiformats/go-multiaddr"
)

var PORT = 10000
var PEERS = []string{
	"QmePbkszdMhjWGPo44meahpHA4noi8w9wrxpFhQkUUbpRg",
	"QmcVUVQijK1kUFYAtifmhMR3SVKfr3u4HRySYBM7Xf86nH",
}

func main() {
	id := flag.Int("id", 0, "peer number to start")
	flag.Parse()

	if *id < 1 {
		panic("id should be greater than 0")
	}

	done := make(chan bool, 1)
	host := makeNode(*id, PORT, done)

	host.run(done)
}

func makeNode(id int, port int, done chan bool) *Node {
	ctx := context.Background()

	// 读取固定的私钥文件
	priv, err := utils.GeneratePrivateKey(fmt.Sprintf("host%d.pem", id))
	if err != nil {
		panic(err)
	}

	// 构建 BasicHost
	listen, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
	basicHost, _ := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrs(listen),
	)
	fmt.Printf("I am %v, please connect to me \n", basicHost.ID())

	// 构建 DHT
	dht, err := kaddht.New(ctx, basicHost)
	if err != nil {
		panic(fmt.Sprintf("new dht failed: err = %v", err))
	}
	routedHost := rhost.Wrap(basicHost, dht)

	// bootstrap the DHT
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