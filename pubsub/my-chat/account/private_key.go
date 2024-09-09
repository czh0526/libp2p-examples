package account

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/czh0526/libp2p-examples/pubsub/my-chat/config"
	libp2p_crypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"os"
	"path/filepath"
)

// createEcdsaPrivateKey 生成一个新的 ECDSA 私钥，
// 并将其转换为 DER 格式，
// 然后转换为 libp2p 的 crypto.PrivKey 类型，并生成对应的 peer ID。
func createEcdsaPrivateKey() ([]byte, string, error) {
	// 生成一个新的 ECDSA 私钥
	ecPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, "", fmt.Errorf("generate ecdsa private key failed, err = %v", err)
	}

	// 将 ECDSA 私钥转换为 DER 格式
	der, err := x509.MarshalECPrivateKey(ecPrivateKey)
	if err != nil {
		return nil, "", fmt.Errorf("marshal ec private key failed, err = %v", err)
	}

	// 将 DER 格式的私钥转换为 libp2p 的 crypto.PrivKey
	privateKey, err := libp2p_crypto.UnmarshalECDSAPrivateKey(der)
	if err != nil {
		return nil, "", fmt.Errorf("convert private key from `ec` to `libp2p` failed, err = %v", err)
	}

	// 生成id
	id, err := peer.IDFromPrivateKey(privateKey)
	if err != nil {
		return nil, "", fmt.Errorf("generate id failed, err = %v", err)
	}

	return der, id.String(), nil
}

// generatePrivateKeyFile 生成一个加密的 PEM 格式的私钥文件，
// 并返回 DER 格式的私钥和对应的 ID。
// 使用给定的密码对私钥进行加密。
func generatePrivateKeyFile(passphrase string) ([]byte, string, error) {
	// 生成`DER格式的私钥`和`id`
	privateKeyDer, id, err := createEcdsaPrivateKey()
	if err != nil {
		return nil, "", fmt.Errorf("create private key failed, err = %v", err)
	}

	// 使用{passphrase}，将`DER格式的私钥`加密为`pem格式的私钥`
	encryptedPrivateKeyPem, err := encryptToPem(privateKeyDer, []byte(passphrase))
	if err != nil {
		return nil, "", fmt.Errorf("encrypt private key failed, err = %v", err)
	}

	// 生成私钥文件
	privateKeyFilename, err := getPrivateKeyFile(id)
	if err != nil {
		return nil, "", fmt.Errorf("create private file failed, err = %v", err)
	}

	// 写入私钥文件
	err = os.WriteFile(privateKeyFilename, encryptedPrivateKeyPem, 0644)
	if err != nil {
		return nil, "", fmt.Errorf("write priv key failed, err = %v", err)
	}

	return privateKeyDer, id, nil
}

// getPrivateKeyFile 根据给定的 ID 生成私钥文件的路径。
func getPrivateKeyFile(id string) (string, error) {
	privateKeyFilename := fmt.Sprintf("%s.pem", id)
	privateKeyPath := filepath.Join(config.HomeDir, config.AccountDir, privateKeyFilename)
	return privateKeyPath, nil
}

// encryptToPem 将 DER 格式的私钥使用 AES 加密算法和给定的密码加密为 PEM 格式的私钥。
func encryptToPem(derBytes, passphrase []byte) ([]byte, error) {
	if len(passphrase) == 0 {
		return pem.EncodeToMemory(
			&pem.Block{
				Type:  "EC PRIVATE KEY",
				Bytes: derBytes,
			},
		), nil
	}

	block, err := aes.NewCipher(deriveKey(passphrase))
	if err != nil {
		return nil, err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCBCEncrypter(block, iv)
	padLen := aes.BlockSize - len(derBytes)%aes.BlockSize
	padding := bytes.Repeat([]byte{byte(padLen)}, padLen)
	derBytes = append(derBytes, padding...)

	ciphertext := make([]byte, len(derBytes))
	stream.CryptBlocks(ciphertext, derBytes)

	encryptedPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "ENCRYPTED EC PRIVATE KEY",
			Bytes: append(iv, ciphertext...),
		},
	)

	return encryptedPEM, nil
}

// decryptFromPem 将加密的 PEM 格式的私钥使用 AES 解密算法和给定的密码解密为 DER 格式的私钥。
func decryptFromPem(encryptedPem []byte, passphrase string) ([]byte, error) {

	encryptedBlock, _ := pem.Decode(encryptedPem)
	encryptedPrivateKey := encryptedBlock.Bytes
	if len(passphrase) == 0 {
		return encryptedPrivateKey, nil
	}

	block, err := aes.NewCipher(deriveKey([]byte(passphrase)))
	if err != nil {
		return nil, err
	}

	iv := encryptedPrivateKey[:aes.BlockSize]
	encryptedPrivateKey = encryptedPrivateKey[aes.BlockSize:]

	stream := cipher.NewCBCDecrypter(block, iv)
	decryptedBytes := make([]byte, len(encryptedPrivateKey))
	stream.CryptBlocks(decryptedBytes, encryptedPrivateKey)

	decryptedBytes = bytes.TrimRight(decryptedBytes, "\x00")
	return decryptedBytes, nil
}

func deriveKey(password []byte) []byte {
	key := make([]byte, 32)
	copy(key, password)
	return key
}

func loadPrivateKey(id string, passphrase string) (libp2p_crypto.PrivKey, error) {
	privateKeyFilename, err := getPrivateKeyFile(id)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(privateKeyFilename)
	if err != nil {
		return nil, fmt.Errorf("read private key failed, err = %v", err)
	}

	privateKeyDer, err := decryptFromPem(content, passphrase)
	if err != nil {
		return nil, fmt.Errorf("decrypt private key failed, err = %v", err)
	}

	privateKey, err := libp2p_crypto.UnmarshalECDSAPrivateKey(privateKeyDer)
	if err != nil {
		return nil, fmt.Errorf("unmarshal priv key failed, err = %v", err)
	}

	return privateKey, nil
}
