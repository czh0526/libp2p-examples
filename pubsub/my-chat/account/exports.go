package account

import (
	"fmt"
	"github.com/czh0526/libp2p-examples/pubsub/my-chat/common"
	"time"
)

func CreateAccount(nickname string, phone string, passphrase string) ([]byte, string, error) {
	settings, err := loadSettings()
	if err != nil {
		return nil, "", err
	}

	privateKeyDer, id, err := generatePrivateKeyFile(passphrase)
	if err != nil {
		return nil, "", err
	}

	settings.Timestamp = common.GetTimestampString(time.Now())
	settings.Data[nickname] = Account{
		Id:    id,
		Phone: phone,
		//Nickname: nickname,
	}

	err = saveSettings(settings)
	if err != nil {
		return nil, "", err
	}

	return privateKeyDer, id, nil
}

func LoadAccount(nickname string, passphrase string) (*Account, error) {
	settings, err := loadSettings()
	if err != nil {
		return nil, fmt.Errorf("load settings failed, err = %v", err)
	}

	account, ok := settings.Data[nickname]
	if !ok {
		return nil, fmt.Errorf("nickname not found in settings")
	}

	_, err = loadPrivateKey(account.Id, passphrase)
	if err != nil {
		return nil, fmt.Errorf("verify passphrase failed, err = %v", err)
	}

	return &account, nil
}
