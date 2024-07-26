package main

import (
	"fmt"
	p2p "github.com/czh0526/libp2p-examples/multipro/proto"
	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"log"
	"sync"
)

const PING_Request = "/ping/pingreq/0.0.1"
const PING_Response = "/ping/pingresp/0.0.1"

type PingProtocol struct {
	node     *Node
	mu       sync.Mutex
	requests map[string]*p2p.PingRequest
}

func NewPingProtocol(node *Node) *PingProtocol {
	p := &PingProtocol{node: node, requests: make(map[string]*p2p.PingRequest)}
	node.SetStreamHandler(PING_Request, p.onPingRequest)
	return p
}

func (p *PingProtocol) onPingRequest(s network.Stream) {
	fmt.Printf("【ping】Read ping request from %s \n", s.Conn().RemotePeer())
	s.Close()
}

func (p *PingProtocol) Ping(peerId peer.ID) bool {
	fmt.Printf("【ping】Plan to send ping to: %s \n", peerId)

	req := &p2p.PingRequest{
		MessageData: p.node.NewMessageData(uuid.New().String(), false),
		Message:     fmt.Sprintf("Ping from %s", p.node.ID()),
	}

	signature, err := p.node.SignProtoMessage(req)
	if err != nil {
		log.Printf("%s: sign ping data failed: err = %v", p.node.ID(), err)
		return false
	}

	req.MessageData.Sign = signature

	p.mu.Lock()
	p.requests[req.MessageData.Id] = req
	p.mu.Unlock()

	ok := p.node.SendProtoMessage(peerId, PING_Request, req)
	if !ok {
		return false
	}

	fmt.Printf("【ping】 Ping to: %s was sent， Message = %s \n",
		peerId, req.Message)
	return true
}
