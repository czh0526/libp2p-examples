package main

import (
	"fmt"
	p2p "github.com/czh0526/libp2p-examples/multipro/proto"
	"github.com/gogo/protobuf/proto"
	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"io"
	"log"
	"sync"
)

const PING_Request = "/ping/pingreq/0.0.1"
const PING_Response = "/ping/pingresp/0.0.1"

type PingProtocol struct {
	node     *Node
	mu       sync.Mutex
	requests map[string]*p2p.PingRequest
	done     chan bool
}

func NewPingProtocol(node *Node, done chan bool) *PingProtocol {
	p := &PingProtocol{node: node, requests: make(map[string]*p2p.PingRequest), done: done}
	node.SetStreamHandler(PING_Request, p.onPingRequest)
	node.SetStreamHandler(PING_Response, p.onPingResponse)
	return p
}

func (p *PingProtocol) onPingRequest(s network.Stream) {
	data := &p2p.PingRequest{}
	buf, err := io.ReadAll(s)
	if err != nil {
		s.Reset()
		log.Printf("Read ping request failed, err = %v", err)
		return
	}
	s.Close()

	err = proto.Unmarshal(buf, data)
	if err != nil {
		log.Printf("Unmarshal ping request failed, err = %v", err)
		return
	}

	log.Printf("%s: Received ping request from %s, Message = %v",
		s.Conn().LocalPeer(), s.Conn().RemotePeer(), data.Message)

	valid := p.node.AuthenticateMessage(data, data.MessageData)
	if !valid {
		log.Println("Authentication failed")
		return
	}

	log.Printf("%s: Sending ping response to %s. Message id: %s",
		s.Conn().LocalPeer(), s.Conn().RemotePeer(), data.MessageData.Id)

	resp := &p2p.PingResponse{
		MessageData: p.node.NewMessageData(data.MessageData.Id, false),
		Message:     fmt.Sprintf("Ping response from %s", p.node.ID()),
	}

	signature, err := p.node.SignProtoMessage(resp)
	if err != nil {
		log.Printf("Sign response failed: err = %v", err)
		return
	}
	resp.MessageData.Sign = signature
	ok := p.node.SendProtoMessage(s.Conn().RemotePeer(), PING_Response, resp)
	if ok {
		log.Printf("%s: Ping response to %s sent.",
			s.Conn().LocalPeer(), s.Conn().RemotePeer())
	}
	p.done <- true
}

func (p *PingProtocol) onPingResponse(s network.Stream) {
	data := &p2p.PingResponse{}
	buf, err := io.ReadAll(s)
	if err != nil {
		s.Reset()
		log.Printf("Read ping response failed, err = %v", err)
		return
	}
	s.Close()

	err = proto.Unmarshal(buf, data)
	if err != nil {
		log.Printf("Unmarshal ping response failed, err = %v", err)
	}

	valid := p.node.AuthenticateMessage(data, data.MessageData)
	if !valid {
		log.Println("Failed to authenticate message")
		return
	}

	p.mu.Lock()
	_, ok := p.requests[data.MessageData.Id]
	if ok {
		delete(p.requests, data.MessageData.Id)
	} else {
		log.Println("Failed to find request data object for response")
		p.mu.Lock()
		return
	}
	p.mu.Unlock()

	log.Printf("%s: Received ping response from %s. Message id: %s. Message: %s",
		s.Conn().LocalPeer(), s.Conn().RemotePeer(), data.MessageData.Id, data.Message)
	p.done <- true
}

func (p *PingProtocol) Ping(host host.Host) bool {
	log.Printf("%s: Send ping to: %s \n", p.node.ID(), host.ID())

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

	ok := p.node.SendProtoMessage(host.ID(), PING_Request, req)
	if !ok {
		return false
	}

	log.Printf("%s: Ping to: %s was sent. Message Id: %s, Message: %s",
		p.node.ID(), host.ID(), req.MessageData.Id, req.Message)
	return true
}
