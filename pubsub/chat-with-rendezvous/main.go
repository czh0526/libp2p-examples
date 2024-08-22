package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"os"
	"sync"
	"time"
)

var (
	topicNameFlag = flag.String("topicName", "applesauce", "name of the topic to join")
)

func main() {
	flag.Parse()

	// 创建本地主机
	ctx := context.Background()
	h, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"))
	if err != nil {
		panic(err)
	}

	// 启动节点发现模块
	go discoverPeers(ctx, h)

	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		panic(fmt.Sprintf("new gossip sub failed, err = %v", err))
	}
	topic, err := ps.Join(*topicNameFlag)
	if err != nil {
		panic(fmt.Sprintf("join topic failed, err = %v", err))
	}
	go streamConsoleTo(ctx, topic)

	sub, err := topic.Subscribe()
	if err != nil {
		panic(fmt.Sprintf("subscribe topic failed, err = %v", err))
	}
	printMessagesFrom(ctx, sub)
}

func discoverPeers(ctx context.Context, h host.Host) {
	kadDHT := initDHT(ctx, h)
	routingDiscovery := drouting.NewRoutingDiscovery(kadDHT)
	dutil.Advertise(ctx, routingDiscovery, *topicNameFlag)

	anyConnected := false
	for !anyConnected {
		fmt.Println("searching for peers...")
		peerChan, err := routingDiscovery.FindPeers(ctx, *topicNameFlag)
		if err != nil {
			panic(err)
		}
		for p := range peerChan {
			if p.ID == h.ID() {
				continue
			}
			err := h.Connect(ctx, p)
			if err != nil {
				fmt.Printf("Failed connecting to peer(`%s`), err = %v\n", p.ID, err)
			} else {
				fmt.Printf("Connected to peer(`%s`)\n", p.ID)
				anyConnected = true
			}
		}
		time.Sleep(10 * time.Second)
	}
	fmt.Println("Peer discovery complete")
}

func initDHT(ctx context.Context, h host.Host) *dht.IpfsDHT {
	kadDHT, err := dht.New(ctx, h)
	if err != nil {
		panic(err)
	}
	if err = kadDHT.Bootstrap(ctx); err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	for _, peerAddr := range BOOTSTRAP_PEERS {
		peerInfo, err := peer.AddrInfoFromP2pAddr(peerAddr)
		if err != nil {
			fmt.Printf("parse bootstrap peer address failed, err = %v \n", err)
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := h.Connect(ctx, *peerInfo); err != nil {
				fmt.Printf("connect bootstrap peer failed, err = %v \n", err)
			} else {
				fmt.Printf("Connected to bootstrap peer %s\n", peerInfo.ID)
			}
		}()
	}
	wg.Wait()

	return kadDHT
}

func streamConsoleTo(ctx context.Context, topic *pubsub.Topic) {
	reader := bufio.NewReader(os.Stdin)
	for {
		s, err := reader.ReadString('\n')
		if err != nil {
			panic(fmt.Sprintf("read input failed, err = %v \n", err))
		}
		if err := topic.Publish(ctx, []byte(s)); err != nil {
			fmt.Printf("### Publish error: %v \n", err)
		}
	}
}

func printMessagesFrom(ctx context.Context, sub *pubsub.Subscription) {
	for {
		m, err := sub.Next(ctx)
		if err != nil {
			panic(fmt.Sprintf("subscription failed, err = %v", err))
		}
		fmt.Println(m.ReceivedFrom, ": ", string(m.Message.Data))
	}
}
