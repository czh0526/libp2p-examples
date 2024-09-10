package main

import (
	"fmt"
	"github.com/czh0526/libp2p-examples/pubsub/my-chat/account"
	"github.com/spf13/cobra"
	"os"
)

var newAccountCmd = &cobra.Command{
	Use:   "new",
	Short: "new an account",
	Run:   newAccount,
}

type NewAccountArgument struct {
	Nickname   string
	Phone      string
	Passphrase string
}

func fetchNewAccountArgs(cmd *cobra.Command) (*NewAccountArgument, error) {

	nickname, _ := cmd.Flags().GetString("nick")
	phone, _ := cmd.Flags().GetString("phone")
	passphrase, _ := cmd.Flags().GetString("passphrase")

	if len(nickname) == 0 {
		return nil, fmt.Errorf("nickname is required")
	}
	if len(phone) == 0 {
		return nil, fmt.Errorf("phone is required")
	}

	return &NewAccountArgument{
		Nickname:   nickname,
		Phone:      phone,
		Passphrase: passphrase,
	}, nil
}

func newAccount(cmd *cobra.Command, _ []string) {
	args, err := fetchNewAccountArgs(cmd)
	if err != nil {
		fmt.Printf("获取参数出错：%v \n", err)
		os.Exit(1)
	}

	_, id, err := account.NewAccount(args.Nickname, args.Phone, args.Passphrase)
	if err != nil {
		fmt.Printf("创建账户出错，%v \n", err)
		os.Exit(1)
	}

	fmt.Printf("创建账户成功，id = %s \n", id)
}
