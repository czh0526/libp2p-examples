package main

import (
	"fmt"
	p2p "github.com/czh0526/libp2p-examples/multipro/proto"
	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"log"
)

const (
	ECHO_Request  = "/echo/echoreq/0.0.1"
	ECHO_Response = "/echo/echoresp/0.0.1"
)

type EchoProtocol struct {
	node     *Node
	requests map[string]*p2p.EchoRequest
	done     chan bool
}

func NewEchoProtocol(node *Node, done chan bool) *EchoProtocol {
	e := EchoProtocol{
		node:     node,
		requests: make(map[string]*p2p.EchoRequest),
		done:     done,
	}
	node.SetStreamHandler(ECHO_Request, e.onEchoRequest)
	node.SetStreamHandler(ECHO_Response, e.onEchoResponse)
	return &e
}

func (e *EchoProtocol) onEchoRequest(s network.Stream) {

}

func (e *EchoProtocol) onEchoResponse(s network.Stream) {

}

func (e *EchoProtocol) Echo(host host.Host) bool {
	log.Printf("%s: Sending echo to: %s", e.node.ID(), host.ID())

	req := &p2p.EchoRequest{
		MessageData: e.node.NewMessageData(uuid.New().String(), false),
		Message:     fmt.Sprintf("Echo from %s", e.node.ID()),
	}

	signature, err := e.node.SignProtoMessage(req)
	if err != nil {
		log.Println("failed to sign echo message")
		return false
	}

	req.MessageData.Sign = signature

	ok := e.node.SendProtoMessage(host.ID(), ECHO_Request, req)
	if !ok {
		return false
	}

	e.requests[req.MessageData.Id] = req
	log.Printf("%s: Echo to: %s was sent, Message Id: %s: Message: %s",
		e.node.ID(), host.ID(), req.MessageData.Id, req.Message)
	return true
}
