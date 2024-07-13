package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"sync"
	"time"
)

func createNode(ctx context.Context, bootstrapPeers []peer.AddrInfo) (host.Host, *dht.IpfsDHT, error) {
	cmgr, err := connmgr.NewConnManager(0, 20, connmgr.WithGracePeriod(5*time.Minute))
	host, err := libp2p.New(
		libp2p.ConnectionManager(cmgr))
	if err != nil {
		return nil, nil, err
	}

	dht, err := dht.New(ctx, host)
	if err != nil {
		return nil, nil, err
	}

	errs := make(chan error, len(bootstrapPeers))
	var wg sync.WaitGroup
	for _, p := range bootstrapPeers {
		wg.Add(1)
		go func(p peer.AddrInfo) {
			defer func() {
				wg.Done()
				fmt.Printf("bootstrap dial: host = %v, peer = %v \n", host.ID(), p.ID)
			}()

			fmt.Printf("%s bootstrapping to %s \n", host.ID(), p.ID)

			host.Peerstore().AddAddrs(p.ID, p.Addrs, peerstore.PermanentAddrTTL)
			fmt.Printf("%s connect to %s \n", host.ID(), p)
			for _, a := range p.Addrs {
				fmt.Printf("\t addr => %s\n", a)
			}
			if err := host.Connect(ctx, p); err != nil {
				fmt.Printf("bootstrapDialFailed %s\n, err = %v \n", p.ID, err)
				errs <- err
				return
			}
		}(p)
	}
	wg.Wait()

	close(errs)
	count := 0
	for err = range errs {
		if err != nil {
			count++
		}
	}
	if count == len(bootstrapPeers) {
		return nil, nil, fmt.Errorf("failed to bootstrap. %s", err)
	}

	return host, dht, nil
}

func connectNodes(ctx context.Context, h1, h2 host.Host) error {
	addrInfo2 := peer.AddrInfo{
		ID:    h2.ID(),
		Addrs: h2.Addrs(),
	}

	return h1.Connect(ctx, addrInfo2)
}

func setupDiscovery(ctx context.Context, h1 host.Host, dht *dht.IpfsDHT) (*routing.RoutingDiscovery, error) {
	routingDiscovery := routing.NewRoutingDiscovery(dht)
	_, err := routingDiscovery.Advertise(ctx, "hello-world")
	if err != nil {
		return nil, err
	}
	return routingDiscovery, nil
}

func main() {
	ctx := context.Background()

	global := flag.Bool("global", false, "use global ipfs peers for bootstrapping")
	flag.Parse()

	var bootstrapPeers []peer.AddrInfo
	if *global {
		fmt.Println("using global bootstrap")
		bootstrapPeers = IPFS_PEERS
	} else {
		fmt.Println("using local bootstrap")
		bootstrapPeers = LOCAL_PEERS
	}

	h1, dht1, err := createNode(ctx, bootstrapPeers)
	if err != nil {
		panic(fmt.Sprintf("create node 1 failed: %v", err))
	}
	defer h1.Close()
	defer dht1.Close()

	h2, dht2, err := createNode(ctx, bootstrapPeers)
	if err != nil {
		panic(fmt.Sprintf("create node 2 failed: %v", err))
	}
	defer h2.Close()
	defer dht2.Close()

	disc1, err := setupDiscovery(ctx, h1, dht1)
	if err != nil {
		panic(fmt.Sprintf("setup discovery 1 failed: %v", err))
	}

	_, err = setupDiscovery(ctx, h2, dht2)
	if err != nil {
		panic(fmt.Sprintf("setup discovery 2 failed: %v", err))
	}

	time.Sleep(5 * time.Second)

	peers, err := disc1.FindPeers(ctx, "hello-world")
	if err != nil {
		panic(fmt.Sprintf("find peers failed: %v", err))
	}

	for p := range peers {
		if p.ID != h1.ID() {
			fmt.Println("Node 1 connecting to Node 2:", p)
			if err := connectNodes(ctx, h1, h2); err != nil {
				panic(err)
			}
			fmt.Println("Node 1 connected to Node 2")
			break
		}
	}

	h2.SetStreamHandler("/chat/1.0.0", func(s network.Stream) {
		fmt.Println("Node 2 received a connection")
		s.Close()
	})

	// 模拟发送数据
	s, err := h1.NewStream(ctx, h2.ID(), "/chat/1.0.0")
	if err != nil {
		panic(err)
	}
	fmt.Println("Node 1 opened a stream to Node 2")
	s.Close()

	select {}
}
