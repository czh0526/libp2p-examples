package account

import (
	"github.com/czh0526/libp2p-examples/pubsub/my-chat/common"
	"github.com/libp2p/go-libp2p/core/crypto"
	"time"
)

func CreateAccount(nickname string, phone string) (crypto.PrivKey, string, error) {
	settings, err := loadSettings()
	if err != nil {
		return nil, "", err
	}

	privateKey, id, err := createPrivateKey()
	if err != nil {
		return nil, "", err
	}

	settings.Timestamp = common.GetTimestampString(time.Now())
	settings.Data[nickname] = Account{
		Id:    id,
		Phone: phone,
	}

	err = saveSettings(settings)
	if err != nil {
		return nil, "", err
	}

	return privateKey, id, nil
}
