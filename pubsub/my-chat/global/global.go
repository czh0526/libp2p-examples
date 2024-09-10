package global

import (
	"fmt"
	"github.com/czh0526/libp2p-examples/pubsub/my-chat/account"
	"github.com/czh0526/libp2p-examples/pubsub/my-chat/links"
	libp2p_crypto "github.com/libp2p/go-libp2p/core/crypto"
)

var (
	myAccount *account.Account
	myFriends map[string]*links.Friend
	myGroups  map[string]*links.Group
)

func GetPrivateKey(passphrase string) (libp2p_crypto.PrivKey, error) {
	if myAccount == nil {
		return nil, fmt.Errorf("account is nil")
	}

	return account.GetPrivateKey(myAccount.Id, passphrase)
}

func GetMyAccount(nickname string) (*account.Account, error) {
	var err error

	if myAccount == nil || myAccount.Nickname != nickname {
		myAccount, err = account.GetAccount(nickname)
		if err != nil {
			return nil, err
		}
	}

	return myAccount, nil
}

func GetMyFriends(myId string) (map[string]*links.Friend, error) {
	var err error

	if myAccount == nil || myAccount.Id != myId {
		return nil, fmt.Errorf("请先登录账号，再加载好友列表")
	}

	if myFriends == nil {
		myFriends, err = links.LoadMyFriends(myId)
		if err != nil {
			return nil, err
		}
	}

	return myFriends, nil
}

func GetMyGroups(myId string) (map[string]*links.Group, error) {
	var err error

	if myAccount == nil || myAccount.Id != myId {
		return nil, fmt.Errorf("请先登录账号，再加载群列表")
	}

	if myGroups == nil {
		myGroups, err = links.LoadMyGroup(myId)
		if err != nil {
			return nil, err
		}
	}

	return myGroups, nil
}
