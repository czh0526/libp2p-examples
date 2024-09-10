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

	if err = createDir(AccountDir); err != nil {
		panic(err)
	}
	if err = createDir(FriendsDir); err != nil {
		panic(err)
	}
	if err = createDir(GroupsDir); err != nil {
		panic(err)
	}
}

func createDir(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(path, os.ModePerm)
			if err != nil {
				return fmt.Errorf("create directory `%s` failed, err = %v", path, err)
			}
		} else {
			return fmt.Errorf("find `%s` failed: err = %s", path, err)
		}
	}

	return nil
}

var (
	HomeDir    string
	AccountDir = "data/account"
	FriendsDir = "data/friends"
	GroupsDir  = "data/groups"
)
