package main

import (
	"context"
	"encoding/json"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

const ChatRoomBufSize = 128

type ChatRoom struct {
	Messages chan *ChatMessage

	ctx   context.Context
	ps    *pubsub.PubSub
	topic *pubsub.Topic
	sub   *pubsub.Subscription

	roomName string
	self     peer.ID
	nick     string
}

type ChatMessage struct {
	Message    string
	SenderID   string
	SenderNick string
}

// 构建一个本地的聊天室镜像
func JoinChatRoom(ctx context.Context, ps *pubsub.PubSub,
	selfID peer.ID, nickname string, roomName string) (*ChatRoom, error) {

	// 加入聊天室主题
	topic, err := ps.Join(topicName(roomName))
	if err != nil {
		return nil, err
	}

	// 订阅聊天室主题
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	chatroom := &ChatRoom{
		ctx:      ctx,
		ps:       ps,
		topic:    topic,
		sub:      sub,
		self:     selfID,
		nick:     nickname,
		roomName: roomName,
	}

	go chatroom.readLoop()
	return chatroom, nil
}

func topicName(roomName string) string {
	return "chat-room:" + roomName
}

func (chatroom *ChatRoom) readLoop() {
	for {
		// 拉取订阅消息
		msg, err := chatroom.sub.Next(chatroom.ctx)
		if err != nil {
			close(chatroom.Messages)
			return
		}

		// 过滤掉自己发出的消息
		if msg.ReceivedFrom == chatroom.self {
			continue
		}

		// 放入消息管道
		cm := new(ChatMessage)
		err = json.Unmarshal(msg.Data, cm)
		if err != nil {
			continue
		}
		chatroom.Messages <- cm
	}
}

func (chatroom *ChatRoom) Publish(message string) error {
	cm := ChatMessage{
		Message:    message,
		SenderID:   chatroom.self.String(),
		SenderNick: chatroom.nick,
	}
	msgBytes, err := json.Marshal(cm)
	if err != nil {
		return err
	}

	return chatroom.topic.Publish(chatroom.ctx, msgBytes)
}
