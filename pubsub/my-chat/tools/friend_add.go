package main

import (
	"fmt"
	"github.com/czh0526/libp2p-examples/pubsub/my-chat/links"
	"github.com/spf13/cobra"
	"os"
)

var addFriendCmd = &cobra.Command{
	Use:   "add",
	Short: "add a friend",
	Run:   addFriend,
}

type AddFriendArgument struct {
	Id       string
	Nickname string
	SelfId   string
}

func fetchAddFriendArgs(cmd *cobra.Command) (*AddFriendArgument, error) {
	nickname, _ := cmd.Flags().GetString("nick")
	id, _ := cmd.Flags().GetString("id")
	selfId, _ := cmd.Flags().GetString("self-id")

	return &AddFriendArgument{
		Id:       id,
		Nickname: nickname,
		SelfId:   selfId,
	}, nil
}

func addFriend(cmd *cobra.Command, _ []string) {
	args, err := fetchAddFriendArgs(cmd)
	if err != nil {
		fmt.Printf("Error parsing arguments: %v\n", err)
		os.Exit(1)
	}

	friends, err := links.LoadMyFriends(args.SelfId)
	if err != nil {
		fmt.Printf("加载好友列表出错，%v \n", err)
		os.Exit(1)
	}

	friend, exists := friends[args.Id]
	if !exists {
		friend = &links.Friend{
			Nickname: args.Nickname,
		}
	}
	friends[args.Id] = friend

	err = links.SaveFriends(args.SelfId, friends)
	if err != nil {
		fmt.Printf("保存好友列表出错，%v \n", err)
		os.Exit(1)
	}

	fmt.Printf("添加好友`%s`成功.\n", args.Nickname)
}
