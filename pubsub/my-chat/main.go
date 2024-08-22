package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"os"
)

func main() {
	nickFlag := flag.String("nick", "", "nickname to use in chat")
	roomFlag := flag.String("room", "awesome-chat-room", "name of chat room to join")
	flag.Parse()
	ctx := context.Background()

	// 创建主机
	h, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"))
	if err != nil {
		panic(fmt.Sprintf("构建host失败，err = %v", err))
	}

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

func shortID(pid peer.ID) string {
	pretty := pid.String()
	return pretty[len(pretty)-8:]
}

func printErr(m string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, m, args...)
}
