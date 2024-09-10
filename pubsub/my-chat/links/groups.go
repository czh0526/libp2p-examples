package links

import (
	db2 "github.com/czh0526/libp2p-examples/pubsub/my-chat/db"
	"github.com/czh0526/libp2p-examples/pubsub/my-chat/db/model"
	"gorm.io/gorm"
)

type Group struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Type  int32  `json:"type"`
	Owner string `json:"owner"`
}

func (g *Group) LoadFromModel(groupModel *model.Group) error {
	g.Id = groupModel.ID
	g.Name = groupModel.Name
	g.Type = groupModel.Type
	g.Owner = groupModel.Owner
	return nil
}

func (g *Group) SaveToGroup() *model.Group {
	return &model.Group{
		ID:    g.Id,
		Name:  g.Name,
		Type:  g.Type,
		Owner: g.Owner,
	}
}

func LoadMyGroup(myId string) (map[string]*Group, error) {
	db, err := db2.GetDB()
	if err != nil {
		return nil, err
	}

	// 读文件
	session := db.Session(&gorm.Session{})
	var groupList []*model.Group
	err = session.Model(groupList).Where("owner = ?", myId).Find(&groupList).Error
	if err != nil {
		return nil, err
	}

	// 解析
	groups := make(map[string]*Group)
	for _, groupModel := range groupList {
		group := &Group{}
		err = group.LoadFromModel(groupModel)
		if err != nil {
			return nil, err
		}
		groups[group.Id] = group
	}

	return groups, nil
}

func JoinGroup(myId string, group *Group) error {
	return nil
}
