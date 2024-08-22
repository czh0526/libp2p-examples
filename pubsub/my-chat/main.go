package main

import (
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
	"log"
	"os"
	"sync"
	"time"
)

var (
	nickFlag = flag.String("nick", "", "nickname to use in chat")
	roomFlag = flag.String("room", "awesome-chat-room", "name of chat room to join")
)

func main() {
	flag.Parse()
	ctx := context.Background()

	// 创建主机
	h, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"))
	if err != nil {
		panic(fmt.Sprintf("构建host失败，err = %v", err))
	}

	// 生成日志
	file, err := os.OpenFile(fmt.Sprintf(
		"chat_%s.log", shortID(h.ID())), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	log.SetOutput(file)

	// 启动节点发现模块
	go discoverPeers(ctx, h)

	// 创建订阅服务
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		panic(fmt.Sprintf("创建GossipSub失败，err = %v", err))
	}

	nickname := *nickFlag
	if len(nickname) == 0 {
		nickname = shortID(h.ID())
	}
	room := *roomFlag

	// 创建聊天室
	chatroom, err := JoinChatRoom(ctx, ps, h.ID(), nickname, room)
	if err != nil {
		panic(fmt.Sprintf("加入聊天室失败，err = %v", err))
	}

	// 创建UI界面
	ui := NewChatUI(chatroom)
	if err = ui.Run(); err != nil {

	}
}

func discoverPeers(ctx context.Context, h host.Host) {
	kadDHT := initDHT(ctx, h)
	routingDiscovery := drouting.NewRoutingDiscovery(kadDHT)
	dutil.Advertise(ctx, routingDiscovery, *roomFlag)

	anyConnected := false
	for !anyConnected {
		log.Println("searching for peers...")
		peerChan, err := routingDiscovery.FindPeers(ctx, *roomFlag)
		if err != nil {
			panic(err)
		}
		for p := range peerChan {
			if p.ID == h.ID() {
				continue
			}
			err := h.Connect(ctx, p)
			if err != nil {
				log.Printf("Failed connecting to peer(`%s`), err = %v\n", p.ID, err)
			} else {
				log.Printf("Connected to peer(`%s`)\n", p.ID)
				anyConnected = true
			}
		}
		time.Sleep(10 * time.Second)
	}
	log.Println("Peer discovery complete")
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
			log.Printf("parse bootstrap peer address failed, err = %v \n", err)
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := h.Connect(ctx, *peerInfo); err != nil {
				log.Printf("connect bootstrap peer failed, err = %v \n", err)
			} else {
				log.Printf("Connected to bootstrap peer %s\n", peerInfo.ID)
			}
		}()
	}
	wg.Wait()

	return kadDHT
}

func shortID(pid peer.ID) string {
	pretty := pid.String()
	return pretty[len(pretty)-8:]
}

func printErr(m string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, m, args...)
}
