package account

import (
	"fmt"
	"github.com/czh0526/libp2p-examples/pubsub/my-chat/config"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"io"
	"os"
	"path/filepath"
)

func getPrivateKeyFile(id string) (io.ReadWriteCloser, error) {
	privateKeyFilename := fmt.Sprintf("%s.pem", id)
	privKeyPath := filepath.Join(config.HomeDir, config.AccountDir, privateKeyFilename)
	return os.OpenFile(privKeyPath, os.O_RDWR|os.O_CREATE, 0600)
}

func createPrivateKey() (crypto.PrivKey, string, error) {
	// 生成私钥
	privateKey, _, err := crypto.GenerateKeyPair(
		crypto.ECDSA, 2048)
	if err != nil {
		return nil, "", fmt.Errorf("create private key failed, err = %v", err)
	}

	// 生成id
	id, err := peer.IDFromPrivateKey(privateKey)
	if err != nil {
		return nil, "", fmt.Errorf("generate id failed, err = %v", err)
	}

	// 生成私钥文件
	privateKeyFile, err := getPrivateKeyFile(id.String())
	if err != nil {
		return nil, "", fmt.Errorf("create private file failed, err = %v", err)
	}

	// 序列化私钥
	privateBytes, err := crypto.MarshalPrivateKey(privateKey)
	if err != nil {
		return nil, "", fmt.Errorf("marshal priv key failed, err = %v", err)
	}

	// 写入私钥文件
	written, err := privateKeyFile.Write(privateBytes)
	if err != nil {
		return nil, "", fmt.Errorf("write priv key failed, err = %v", err)
	}
	if written != len(privateBytes) {
		return nil, "", fmt.Errorf("write priv key not completed")
	}

	return privateKey, id.String(), nil
}

func loadPrivateKey(id string) (crypto.PrivKey, error) {
	privKeyFile, err := getPrivateKeyFile(id)
	if err != nil {
		return nil, err
	}
	defer privKeyFile.Close()

	privateBytes, err := io.ReadAll(privKeyFile)
	if err != nil {
		return nil, fmt.Errorf("read priv key failed, err = %v", err)
	}

	privateKey, err := crypto.UnmarshalPrivateKey(privateBytes)
	if err != nil {
		return nil, fmt.Errorf("unmarshal priv key failed, err = %v", err)
	}

	return privateKey, nil
}
