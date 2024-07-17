package utils

import (
	"fmt"
	"github.com/libp2p/go-libp2p/core/crypto"
	"os"
)

func GeneratePrivateKey(filename string) (crypto.PrivKey, error) {
	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			privateKey, _, err := crypto.GenerateKeyPair(
				crypto.ECDSA, 2048)
			if err != nil {
				return nil, fmt.Errorf("create priv key failed, err = %v", err)
			}
			privateBytes, err := crypto.MarshalPrivateKey(privateKey)
			if err != nil {
				return nil, fmt.Errorf("marshal priv key failed, err = %v", err)
			}
			err = os.WriteFile(filename, privateBytes, 0600)
			if err != nil {
				return nil, fmt.Errorf("write priv key failed, err = %v", err)
			}

			return privateKey, nil
		}
		return nil, err
	}

	privateBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read priv key failed, err = %v", err)
	}

	privateKey, err := crypto.UnmarshalPrivateKey(privateBytes)
	if err != nil {
		return nil, fmt.Errorf("unmarshal priv key failed, err = %v", err)
	}

	return privateKey, nil
}
