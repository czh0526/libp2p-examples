package friend

import (
	"fmt"
	"github.com/czh0526/libp2p-examples/pubsub/my-chat/config"
	"path/filepath"
)

type Friend struct {
	Id       string `json:"id"`
	Nickname string `json:"nickname"`
	GroupId  string `json:"groupId"`
}

func getFriendsFile(id string) (string, error) {
	filename := fmt.Sprintf("%s.json", id)
	path := filepath.Join(config.HomeDir, config.FriendsDir, filename)
	return path, nil
}

func LoadFriends(selfId string) (map[string]Friend, error) {
	friendsFilename, err := getFriendsFile(selfId)

}
