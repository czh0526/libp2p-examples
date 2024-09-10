package links

import (
	"errors"
	"fmt"
	db2 "github.com/czh0526/libp2p-examples/pubsub/my-chat/db"
	"github.com/czh0526/libp2p-examples/pubsub/my-chat/db/model"
	"gorm.io/gorm"
)

type Friend struct {
	Id       string `json:"id"`
	Nickname string `json:"nickname"`
	Owner    string `json:"owner"`
}

func (f *Friend) LoadFromModel(friendModel *model.Friend) error {
	f.Id = friendModel.ID
	f.Nickname = friendModel.Nickname
	f.Owner = friendModel.Owner
	return nil
}

func (f *Friend) SaveToModel() *model.Friend {
	return &model.Friend{
		ID:       f.Id,
		Nickname: f.Nickname,
		Owner:    f.Owner,
	}
}

func LoadMyFriends(myId string) (map[string]*Friend, error) {
	db, err := db2.GetDB()
	if err != nil {
		return nil, err
	}

	// 读文件
	session := db.Session(&gorm.Session{})
	var friendList []*model.Friend
	err = session.Model(friendList).Where("owner = ?", myId).Find(&friendList).Error
	if err != nil {
		return nil, err
	}

	// 解析
	friends := make(map[string]*Friend)
	for _, friendModel := range friendList {
		friend := &Friend{}
		err = friend.LoadFromModel(friendModel)
		if err != nil {
			return nil, err
		}
		friends[friend.Id] = friend
	}

	return friends, nil
}

func GetFriend(id string) (*Friend, error) {
	db, err := db2.GetDB()
	if err != nil {
		return nil, err
	}

	session := db.Session(&gorm.Session{})
	var friendModel model.Friend
	err = session.Where("id = ?", id).First(&friendModel).Error
	if err != nil {
		return nil, err
	}

	friend := &Friend{}
	_ = friend.LoadFromModel(&friendModel)

	return friend, nil
}

func AddFriend(myId string, friend *Friend) error {
	friend, err := GetFriend(friend.Id)
	if err == nil {
		return fmt.Errorf("%s(`%s`) has already be added", friend.Nickname, friend.Id)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("find `%s` failed, err = %v", friend.Nickname, err)
	}

	friend.Owner = myId
	friendModel := friend.SaveToModel()
	err = createFriend(friendModel)
	if err != nil {
		return err
	}

	return nil
}

func createFriend(friendModel *model.Friend) error {
	db, err := db2.GetDB()
	if err != nil {
		return err
	}

	tx := db.Begin()
	if err = tx.Create(friendModel).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
