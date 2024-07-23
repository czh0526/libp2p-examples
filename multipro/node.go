package main

import (
	"context"
	"fmt"
	p2p "github.com/czh0526/libp2p-examples/multipro/proto"
	ggio "github.com/gogo/protobuf/io"
	"github.com/gogo/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"log"
	"time"
)

const (
	clientVersion = "go-p2p-node/0.0.1"
)

type Node struct {
	host.Host
	*PingProtocol
	*EchoProtocol
}

func NewNode(host host.Host, done chan bool) *Node {
	node := &Node{Host: host}
	node.PingProtocol = NewPingProtocol(node, done)
	node.EchoProtocol = NewEchoProtocol(node, done)
	return node
}

func (n *Node) run(done <-chan bool) {
	myId := n.ID()
	for _, pid := range PEERS {
		peerId := peer.ID(pid)
		if peerId == myId {
			continue
		}

		if !n.Ping(peerId) {
			fmt.Printf("Peer %s is down\n", peerId)
		}

		time.Sleep(10 * time.Second)
	}

	//h1.Echo(h2.Host)
	//h2.Echo(h1.Host)

	select {}
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
