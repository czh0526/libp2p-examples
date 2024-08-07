package main

import (
	"fmt"
	p2p "github.com/czh0526/libp2p-examples/multipro/proto"
	"github.com/gogo/protobuf/proto"
	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/network"
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
}

func NewEchoProtocol(node *Node) *EchoProtocol {
	e := EchoProtocol{
		node:     node,
		requests: make(map[string]*p2p.EchoRequest),
	}
	node.SetStreamHandler(ECHO_Request, e.onEchoRequest)
	return &e
}

func (e *EchoProtocol) onEchoRequest(s network.Stream) {
	defer s.Close()
	fmt.Printf("【echo】Read `echo` request from %s \n", s.Conn().RemotePeer())

	data := &p2p.EchoRequest{}
	buf, err := io.ReadAll(s)
	if err != nil {
		s.Reset()
		log.Println(err)
		return
	}
	fmt.Printf("【echo】Read `echo` data %v bytes \n", len(buf))

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

	fmt.Printf("【echo】verify echo response successfully, Message = %v\n",
		data.Message)
}

func (e *EchoProtocol) Echo(s network.Stream) bool {
	defer s.Close()
	fmt.Printf("【echo】 Plan to send echo to: %s. \n", s.Conn().RemotePeer())

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

	ok := e.node.SendProtoMessage(s, req)
	if !ok {
		return false
	}

	e.requests[req.MessageData.Id] = req
	fmt.Printf("【echo】 Echo to: %s was sent, Message: %s \n", s.Conn().RemotePeer(), req.Message)

	return true
}
