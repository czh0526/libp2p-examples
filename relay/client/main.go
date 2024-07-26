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

var PEERS = []string{
	"QmePbkszdMhjWGPo44meahpHA4noi8w9wrxpFhQkUUbpRg",
	"QmcVUVQijK1kUFYAtifmhMR3SVKfr3u4HRySYBM7Xf86nH",
}

var (
	RELAY_ENDPOINT  = "/ip4/9.134.4.207/tcp/8000/p2p/QmfNuQPFFuqw6x2cptzRwmnZah1hJBdQ3niTBLSEpJKgmd"
	RELAY_ADDR_INFO = convertPeer(RELAY_ENDPOINT)
)

func main() {
	id := flag.Int("id", 0, "peer number to start")
	flag.Parse()

	if *id < 1 {
		panic("id should be greater than 0")
	}

	host := makeNode(*id)

	host.run()
}

func makeNode(id int) *Node {
	ctx := context.Background()

	// 读取固定的私钥文件
	priv, err := utils.GeneratePrivateKey(fmt.Sprintf("host%d.pem", id))
	if err != nil {
		panic(err)
	}

	// 构建 BasicHost
	listen, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/10000"))
	basicHost, _ := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrs(listen),
		//libp2p.NoListenAddrs,
		libp2p.EnableRelay(), // it's important !!!
	)
	fmt.Printf("I am %v \n", basicHost.ID())
	fmt.Printf("I am listening on %v \n", listen)

	// 构建 DHT
	dht, err := kaddht.New(ctx, basicHost)
	if err != nil {
		panic(fmt.Sprintf("new dht failed: err = %v", err))
	}
	routedHost := rhost.Wrap(basicHost, dht)

	// 启动 DHT 服务
	err = dht.Bootstrap(ctx)
	if err != nil {
		panic(fmt.Sprintf("host bootstrap failed, err = %v", err))
	}

	// 连接 Bootstrap 节点
	err = bootstrapConnect(ctx, routedHost, BOOTSTRAP_PEERS)
	if err != nil {
		panic(fmt.Sprintf("connect bootstrap peers failed, err = %v", err))
	}

	return NewNode(routedHost)
}
