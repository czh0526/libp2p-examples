package config

import (
	"fmt"
	"os"
)

func init() {
	pwd, err := os.Getwd()
	if err != nil {
		panic("cannot get current working directory")
	}
	HomeDir = pwd

	_, err = os.Stat(AccountDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(AccountDir, os.ModePerm)
			if err != nil {
				panic("cannot create account directory")
			}
		} else {
			panic(fmt.Sprintf("find `%s` failed: err = %s", AccountDir, err))
		}
	}
}

var (
	HomeDir    string
	AccountDir = "data/account"
	FriendsDir = "data/friends"
)
