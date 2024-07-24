package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/czh0526/libp2p-examples/utils"
	"github.com/libp2p/go-libp2p"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	rhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	"github.com/libp2p/go-libp2p/p2p/net/swarm"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
	ma "github.com/multiformats/go-multiaddr"
	"log"
	"time"
)

var (
	RELAY_ADDR = convertPeer(
		"/ip4/9.134.4.207/tcp/8000/p2p/QmfNuQPFFuqw6x2cptzRwmnZah1hJBdQ3niTBLSEpJKgmd")
)

var (
	PEERS = []string{
		"QmePbkszdMhjWGPo44meahpHA4noi8w9wrxpFhQkUUbpRg",
		"QmcVUVQijK1kUFYAtifmhMR3SVKfr3u4HRySYBM7Xf86nH",
	}
)

func convertPeer(addr string) peer.AddrInfo {

	maddr := ma.StringCast(addr)
	p, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		panic(fmt.Sprintf("parse ipfs bootstrap peers failed: err = %v", err))
	}
	return *p
}

func main() {
	id := flag.Int("id", 0, "peer number to start")
	flag.Parse()

	if *id < 1 {
		panic("id should be greater than 0")
	}

	rhost, err := makeHost(context.Background(), fmt.Sprintf("host%d.pem", *id))
	if err != nil {
		panic(fmt.Sprintf("make host failed: err = %v", err))
	}

	rhost.SetStreamHandler("/customprotocol", func(s network.Stream) {
		log.Println("Awesome! we're now communicating via the relay!")
		s.Close()
	})

	for {
		for _, p := range PEERS {
			pid, err := peer.Decode(p)
			if err != nil {
				fmt.Printf("decode pid id(`%s`) failed: err = %v\n", p, err)
			}

			if pid == rhost.ID() {
				continue
			}

			_, err = rhost.NewStream(context.Background(), pid, "/customprotocol")
			if err != nil {
				log.Printf("two hosts connect directly failed, err = %v", err)
				continue
			}

			fmt.Println("【normal】As suspected, we cannot directly dial between two hosts")

			if err := rhost.Connect(context.Background(), RELAY_ADDR); err != nil {
				log.Printf("Failed to connect host and relay: err = %v", err)
				continue
			}

			reservation, err := client.Reserve(context.Background(), rhost, RELAY_ADDR)
			if err != nil {
				log.Printf("host failed to receive a relay reservation from relay, err = %v", err)
				continue
			}
			fmt.Printf("【normal】Reservation = %v\n", reservation)

			relayaddr, err := ma.NewMultiaddr(fmt.Sprintf("/p2p/%s/p2p-circuit/p2p/%s",
				RELAY_ADDR.ID.String(), pid.String()))
			if err != nil {
				log.Println(err)
				continue
			}

			rhost.Network().(*swarm.Swarm).Backoff().Clear(pid)
			fmt.Println("【normal】Now let's attempt to connect the hosts via the relay node")

			relayInfo := peer.AddrInfo{
				ID:    pid,
				Addrs: []ma.Multiaddr{relayaddr},
			}
			if err := rhost.Connect(context.Background(), relayInfo); err != nil {
				log.Printf("Unexpected error here. Failed to connect host1 and host2, err = %v", err)
				continue
			}

			fmt.Println("【normal】Yep, that worked!")

			s, err := rhost.NewStream(
				network.WithAllowLimitedConn(context.Background(), "customprotocol"),
				pid, "/customprotocol")
			if err != nil {
				log.Printf("Whoops, this should have worked..., err = %v \n", err)
				continue
			}

			s.Read(make([]byte, 1))
		}
		time.Sleep(10 * time.Second)
	}
}

func makeHost(ctx context.Context, keyFilename string) (host.Host, error) {
	priv, err := utils.GeneratePrivateKey(keyFilename)
	if err != nil {
		panic(err)
	}

	listen, _ := ma.NewMultiaddr("/ip4/0.0.0.0/tcp/10000")
	basicHost, err := libp2p.New(
		libp2p.Identity(priv),
		//libp2p.NoListenAddrs,
		libp2p.ListenAddrs(listen),
		libp2p.EnableRelay(),
	)
	if err != nil {
		log.Printf("Failed to create host, err = %v", err)
		return nil, err
	}
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

	return basicHost, nil
}
