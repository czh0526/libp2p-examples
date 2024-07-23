package main

import (
	"fmt"
	p2p "github.com/czh0526/libp2p-examples/multipro/proto"
	"github.com/gogo/protobuf/proto"
	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"io"
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
	data := &p2p.EchoRequest{}
	buf, err := io.ReadAll(s)
	if err != nil {
		s.Reset()
		log.Println(err)
		return
	}
	s.Close()

	err = proto.Unmarshal(buf, data)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("【echo】Received echo request from %s, Message = %v\n",
		s.Conn().RemotePeer().String(), data.Message)

	valid := e.node.AuthenticateMessage(data, data.MessageData)
	if !valid {
		log.Println("Failed to authenticate message")
		return
	}

	fmt.Printf("【echo】Sending echo response to %s, Message = %v\n",
		s.Conn().RemotePeer().String(), data.Message)
	// create echo response
	resp := &p2p.EchoResponse{
		Message:     data.Message,
		MessageData: e.node.NewMessageData(data.MessageData.Id, false),
	}
	signature, err := e.node.SignProtoMessage(resp)
	if err != nil {
		log.Printf("failed to sign response, err = %v", err)
		return
	}
	resp.MessageData.Sign = signature

	// send echo response
	ok := e.node.SendProtoMessage(s.Conn().RemotePeer(), ECHO_Response, resp)
	if !ok {
		fmt.Printf("【echo】Echo response to %s sent.", s.Conn().RemotePeer())
	}
	e.done <- true
}

func (e *EchoProtocol) onEchoResponse(s network.Stream) {
	data := &p2p.EchoResponse{}
	buf, err := io.ReadAll(s)
	if err != nil {
		s.Reset()
		log.Println(err)
		return
	}
	s.Close()

	err = proto.Unmarshal(buf, data)
	if err != nil {
		log.Println(err)
		return
	}

	valid := e.node.AuthenticateMessage(data, data.MessageData)
	if !valid {
		log.Printf("Failed to authenticate message, err = %v", err)
		return
	}

	req, ok := e.requests[data.MessageData.Id]
	if ok {
		delete(e.requests, data.MessageData.Id)
	} else {
		log.Printf("Failed to find request for id = %v", data.MessageData.Id)
		return
	}

	fmt.Printf("【echo】Received echo response from %s, Message = %v\n",
		s.Conn().RemotePeer().String(), req.Message)
	e.done <- true
}

func (e *EchoProtocol) Echo(peerId peer.ID) bool {
	log.Printf("【echo】 Plan to send echo to: %s", peerId)

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

	ok := e.node.SendProtoMessage(peerId, ECHO_Request, req)
	if !ok {
		return false
	}

	e.requests[req.MessageData.Id] = req
	fmt.Printf("【echo】 Echo to: %s was sent, Message: %s \n", peerId, req.Message)

	return true
}
