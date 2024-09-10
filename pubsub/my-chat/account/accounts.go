package account

import (
	"fmt"
	db2 "github.com/czh0526/libp2p-examples/pubsub/my-chat/db"
	"github.com/czh0526/libp2p-examples/pubsub/my-chat/db/model"
	libp2p_crypto "github.com/libp2p/go-libp2p/core/crypto"
	"gorm.io/gorm"
)

type Account struct {
	Id       string `json:"id"`
	Phone    string `json:"phone"`
	Nickname string `json:"nickname,omitempty"`
}

func (a *Account) LoadFromModel(accountModel *model.Account) error {
	a.Id = accountModel.ID
	a.Phone = accountModel.Phone
	a.Nickname = accountModel.Nickname

	return nil
}

func (a *Account) SaveToModel() *model.Account {
	return &model.Account{
		ID:       a.Id,
		Phone:    a.Phone,
		Nickname: a.Nickname,
	}
}

func LoadAccounts() (map[string]*Account, error) {
	db, err := db2.GetDB()
	if err != nil {
		return nil, err
	}

	// 读文件
	session := db.Session(&gorm.Session{})
	var accounts []*model.Account
	if err = session.Find(&accounts).Error; err != nil {
		return nil, err
	}

	// 解析json
	accountsMap := make(map[string]*Account)
	for _, accountModel := range accounts {
		account := &Account{}
		_ = account.LoadFromModel(accountModel)
		accountsMap[account.Id] = account
	}
	return accountsMap, nil
}

func NewAccount(nickname string, phone string, passphrase string) ([]byte, string, error) {
	accounts, err := LoadAccounts()
	if err != nil {
		return nil, "", err
	}

	_, exists := accounts[nickname]
	if exists {
		return nil, "", fmt.Errorf("nickname %s already exists", nickname)
	}

	privateKeyDer, id, err := generatePrivateKeyFile(passphrase)
	if err != nil {
		return nil, "", err
	}

	accountModel := (&Account{
		Id:       id,
		Phone:    phone,
		Nickname: nickname,
	}).SaveToModel()
	err = createAccount(accountModel)
	if err != nil {
		return nil, "", err
	}

	return privateKeyDer, id, nil
}

func GetAccount(nickname string) (*Account, error) {
	db, err := db2.GetDB()
	if err != nil {
		return nil, err
	}

	session := db.Session(&gorm.Session{})
	accountModel := &model.Account{}
	err = session.Where("nickname = ?", nickname).First(accountModel).Error
	if err != nil {
		return nil, err
	}

	account := &Account{}
	_ = account.LoadFromModel(accountModel)

	return account, nil
}

func GetPrivateKey(accountId string, passphrase string) (libp2p_crypto.PrivKey, error) {
	return loadPrivateKey(accountId, passphrase)
}

func createAccount(accountModel *model.Account) error {
	db, err := db2.GetDB()
	if err != nil {
		return err
	}

	tx := db.Begin()
	if err = tx.Create(accountModel).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
