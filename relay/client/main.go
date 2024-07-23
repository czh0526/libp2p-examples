package main

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/net/swarm"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
	ma "github.com/multiformats/go-multiaddr"
	"log"
)

var (
	RELAY_ADDR = convertPeer(
		"/ip4/9.134.4.207/tcp/8080/QmWiG7ExhxNokqzghHrxC25m3W8gVEftgcrZsJKhPv1Y74")
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
	unreachable1, err := libp2p.New(
		libp2p.NoListenAddrs,
		libp2p.EnableRelay(),
	)
	if err != nil {
		log.Printf("Failed to create unreachable1, err = %v", err)
		return
	}

	unreachable2, err := libp2p.New(
		libp2p.NoListenAddrs,
		libp2p.EnableRelay(),
	)
	if err != nil {
		log.Printf("Failed to create unreachable2, err = %v", err)
		return
	}

	fmt.Println("First let's attempt to directly connect")
	unreachable2Info := peer.AddrInfo{
		ID:    unreachable2.ID(),
		Addrs: unreachable2.Addrs(),
	}
	err = unreachable2.Connect(context.Background(), unreachable2Info)
	if err == nil {
		log.Printf("This actually should have failed.")
		return
	}

	log.Println("As suspected, we cannot directly dial between the unreachable hosts")

	if err := unreachable1.Connect(context.Background(), RELAY_ADDR); err != nil {
		log.Printf("Failed to connect unreachable1 and relay1: err = %v", err)
		return
	}
	if err := unreachable2.Connect(context.Background(), RELAY_ADDR); err != nil {
		log.Printf("Failed to connect unreachable2 and relay1: err = %v", err)
	}

	unreachable2.SetStreamHandler("/customprotocol", func(s network.Stream) {
		log.Println("Awesome! we're now communicating via the relay!")
		s.Close()
	})

	_, err = client.Reserve(context.Background(), unreachable2, RELAY_ADDR)
	if err != nil {
		log.Printf("unreachable2 failed to receive a relay reservation from relay1, err = %v", err)
		return
	}

	relayaddr, err := ma.NewMultiaddr(fmt.Sprintf("/p2p/%s/p2p-circuit/p2p/%s",
		RELAY_ADDR.ID.String(), unreachable2.ID().String()))
	if err != nil {
		log.Println(err)
		return
	}

	unreachable1.Network().(*swarm.Swarm).Backoff().Clear(unreachable2.ID())
	fmt.Println("Now let's attempt to connect the hosts via the relay node")

	unreachable2RelayInfo := peer.AddrInfo{
		ID:    unreachable2.ID(),
		Addrs: []ma.Multiaddr{relayaddr},
	}
	if err := unreachable1.Connect(context.Background(), unreachable2RelayInfo); err != nil {
		log.Printf("Unexpected error here. Failed to connect unreachable1 and unreachable2, err = %v", err)
		return
	}

	fmt.Println("Yep, that worked!")

	s, err := unreachable1.NewStream(
		network.WithAllowLimitedConn(context.Background(), "customprotocol"),
		unreachable2.ID(),
		"/customprotocol")
	if err != nil {
		log.Printf("Whoops, this should have worked..., err = %v \n", err)
		return
	}

	s.Read(make([]byte, 1))
}
