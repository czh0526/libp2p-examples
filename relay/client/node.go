package main

import (
	"context"
	"fmt"
	p2p "github.com/czh0526/libp2p-examples/multipro/proto"
	ggio "github.com/gogo/protobuf/io"
	"github.com/gogo/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/net/swarm"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
	maddr "github.com/multiformats/go-multiaddr"
	"log"
	"time"
)

const (
	clientVersion = "go-p2p-node/0.0.1"
)

type Node struct {
	host.Host
	*PingProtocol
}

func NewNode(host host.Host) *Node {
	node := &Node{Host: host}
	node.PingProtocol = NewPingProtocol(node)

	return node
}

func (n *Node) run() {
	myId := n.ID()
	for {
		for _, pid := range PEERS {
			peerId, err := peer.Decode(pid)
			if err != nil {
				fmt.Printf("decode peer id(`%s`) failed: err = %v\n", pid, err)
			}
			if peerId == myId {
				continue
			}

			if !n.Ping(peerId) {
				fmt.Printf("【ping】`%s` is unreachable \n", peerId)
			} else {
				fmt.Printf("【ping】`%s` is connected \n", peerId)
			}

			if err = n.ConnectByRelay(peerId); err != nil {
				fmt.Println("\n========== bad result ===========")
				fmt.Println()
			}

			time.Sleep(10 * time.Second)
		}
	}
}

func (n *Node) NewMessageData(messageId string, gossip bool) *p2p.MessageData {
	nodePubKey, err := crypto.MarshalPublicKey(n.Peerstore().PubKey(n.ID()))
	if err != nil {
		panic("Failed to marshal public key for sender from local peer store.")
	}

	return &p2p.MessageData{
		ClientVersion: clientVersion,
		NodeId:        n.ID().String(),
		NodePubKey:    nodePubKey,
		Timestamp:     time.Now().Unix(),
		Id:            messageId,
		Gossip:        gossip,
	}
}

func (n *Node) SignProtoMessage(message proto.Message) ([]byte, error) {
	data, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}
	return n.signData(data)
}

func (n *Node) signData(data []byte) ([]byte, error) {
	key := n.Peerstore().PrivKey(n.ID())
	return key.Sign(data)
}

func (n *Node) AuthenticateMessage(message proto.Message, data *p2p.MessageData) bool {
	sign := data.Sign
	data.Sign = nil
	bin, err := proto.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return false
	}
	data.Sign = sign

	peerId, err := peer.Decode(data.NodeId)
	if err != nil {
		log.Printf("Failed to decode peer id from base58: %v", err)
		return false
	}

	return n.verifyData(bin, sign, peerId, data.NodePubKey)
}

func (n *Node) verifyData(data []byte, signature []byte, peerId peer.ID, pubKeyData []byte) bool {
	key, err := crypto.UnmarshalPublicKey(pubKeyData)
	if err != nil {
		log.Println("Failed to unmarshal public key")
		return false
	}

	idFromKey, err := peer.IDFromPublicKey(key)
	if err != nil {
		log.Println("Failed to extract peer id from public key")
		return false
	}

	if idFromKey != peerId {
		log.Println("Invalid peer id")
		return false
	}

	res, err := key.Verify(data, signature)
	if err != nil {
		log.Println("Failed to verify signature")
		return false
	}
	return res
}

func (n *Node) SendProtoMessage(id peer.ID, p protocol.ID, data proto.Message) bool {
	s, err := n.NewStream(context.Background(), id, p)
	if err != nil {
		log.Println(err)
		return false
	}
	defer s.Close()

	writer := ggio.NewFullWriter(s)
	err = writer.WriteMsg(data)
	if err != nil {
		log.Println(err)
		s.Reset()
		return false
	}
	return true
}

func (n *Node) ConnectByRelay(pid peer.ID) error {
	rHost := n.Host

	// 连接 Relay 节点
	if err := rHost.Connect(context.Background(), RELAY_ADDR_INFO); err != nil {
		log.Printf("Failed to connect host and relay: err = %v", err)
		return err

	}

	// 请求`relay节点`预留 slot
	reservation, err := client.Reserve(context.Background(), rHost, RELAY_ADDR_INFO)
	if err != nil {
		log.Printf("host failed to receive a relay reservation from relay, err = %v", err)
		return err
	}
	fmt.Println("【normal】Reservation success")
	fmt.Printf("\t => Expiration = %s\n", reservation.Expiration)
	for _, addr := range reservation.Addrs {
		fmt.Printf("\t=> addr = %s \n", addr)
	}

	// 创建 Relay 地址
	relayAddr, err := maddr.NewMultiaddr(
		fmt.Sprintf("%s/p2p-circuit/p2p/%s", RELAY_ENDPOINT, pid.String()))
	if err != nil {
		log.Printf("create relay address failed, err = %v", err)
		return err
	}

	rHost.Network().(*swarm.Swarm).Backoff().Clear(pid)

	// 创建 Relay AddrInfo
	peerRelayInfo := peer.AddrInfo{
		ID:    pid,
		Addrs: []maddr.Multiaddr{relayAddr},
	}
	fmt.Println("【normal】create AddrInfo for relay link success")
	fmt.Printf("\t => id = %s \n", peerRelayInfo.ID)
	for _, addr := range peerRelayInfo.Addrs {
		fmt.Printf("\t=> addr = %s \n", addr)
	}

	if err := rHost.Connect(context.Background(), peerRelayInfo); err != nil {
		log.Printf("Unexpected error here. Failed to connect host1 with host2: %v", err)
		return err
	}
	log.Println("Yep, that worked!")

	// New Stream
	s, err := rHost.NewStream(
		network.WithAllowLimitedConn(context.Background(), "ping"),
		pid, PING_Request)
	if err != nil {
		log.Printf("Unexpected error here. Failed to new stream between host1 and host2, err = %v", err)
		return err
	}

	defer s.Close()
	fmt.Println("【normal】Yep, that worked!")

	return nil
}
