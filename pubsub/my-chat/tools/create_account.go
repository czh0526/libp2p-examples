package main

import (
	"flag"
	"fmt"
	"github.com/czh0526/libp2p-examples/pubsub/my-chat/account"
	"os"
)

var (
	nicknameFlag = flag.String("nick", "", "nickname to use in chat")
	phoneFlag    = flag.String("phone", "", "phone to use in chat")
)

type Argument struct {
	Nickname string
	Phone    string
}

func getArguments() (*Argument, error) {
	flag.Parse()
	nickname := *nicknameFlag
	phone := *phoneFlag
	if len(nickname) == 0 {
		return nil, fmt.Errorf("nickname is required")
	}
	if len(phone) == 0 {
		return nil, fmt.Errorf("phone is required")
	}

	return &Argument{
		Nickname: nickname,
		Phone:    phone,
	}, nil
}

func main() {
	args, err := getArguments()
	if err != nil {
		fmt.Printf("获取参数出错：%v \n", err)
		os.Exit(1)
	}

	_, id, err := account.CreateAccount(args.Nickname, args.Phone)
	if err != nil {
		fmt.Printf("创建账户出错，%v \n", err)
		os.Exit(1)
	}

	fmt.Printf("创建账户成功，id = %s \n", id)
}
