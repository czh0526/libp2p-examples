package account

import (
	"crypto/x509"
	"fmt"
	libp2p_crypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestCreateEcPrivateKey(t *testing.T) {
	// 生成私钥
	privateKeyDer, id, err := createEcdsaPrivateKey()
	assert.NoError(t, err)
	fmt.Println(string(privateKeyDer))
	fmt.Printf("id = %s\n", id)
}

func TestGeneratePrivateKeyFile_WithPassword(t *testing.T) {
	_, id, err := generatePrivateKeyFile("123456")
	assert.NoError(t, err)

	privateKeyFilename, err := getPrivateKeyFile(id)
	assert.NoError(t, err)

	info, err := os.Stat(privateKeyFilename)
	assert.NoError(t, err)
	fmt.Println(info.Size())
}

func TestGeneratePrivateKeyFile_NoPassword(t *testing.T) {
	_, id, err := generatePrivateKeyFile("")
	assert.NoError(t, err)

	privateKeyFilename, err := getPrivateKeyFile(id)
	assert.NoError(t, err)

	info, err := os.Stat(privateKeyFilename)
	assert.NoError(t, err)
	fmt.Println(info.Size())
}

func TestGeneratePrivateKeyContent_WithPassword(t *testing.T) {
	privateKeyDer, id, err := generatePrivateKeyFile("123456")
	assert.NoError(t, err)
	privateKey, err := x509.ParseECPrivateKey(privateKeyDer)
	assert.NoError(t, err)
	fmt.Printf("id = %s\n", id)

	// 加载私钥文件
	privateKeyFilename, err := getPrivateKeyFile(id)
	assert.NoError(t, err)
	privateKeyPem, err := os.ReadFile(privateKeyFilename)
	assert.NoError(t, err)

	privateKeyDer2, err := decryptFromPem(privateKeyPem, "123456")
	assert.NoError(t, err)
	assert.Equal(t, privateKeyDer2[:len(privateKeyDer)], privateKeyDer)

	privateKey2, err := x509.ParseECPrivateKey(privateKeyDer2)
	assert.NoError(t, err)
	assert.Equal(t, privateKey, privateKey2)
}

func TestGeneratePrivateKeyContent_NoPassword(t *testing.T) {
	privateKeyDer, id, err := generatePrivateKeyFile("")
	assert.NoError(t, err)
	privateKey, err := x509.ParseECPrivateKey(privateKeyDer)
	assert.NoError(t, err)
	fmt.Printf("id = %s\n", id)

	// 加载私钥文件
	privateKeyFilename, err := getPrivateKeyFile(id)
	assert.NoError(t, err)
	privateKeyPem, err := os.ReadFile(privateKeyFilename)
	assert.NoError(t, err)

	privateKeyDer2, err := decryptFromPem(privateKeyPem, "")
	assert.NoError(t, err)
	assert.Equal(t, privateKeyDer2[:len(privateKeyDer)], privateKeyDer)

	privateKey2, err := x509.ParseECPrivateKey(privateKeyDer2)
	assert.NoError(t, err)
	assert.Equal(t, privateKey, privateKey2)
}

func TestLoadPrivateKey_WithPassword(t *testing.T) {

	privateKeyDer, id, err := generatePrivateKeyFile("123456")
	assert.NoError(t, err)
	fmt.Printf("id = %s\n", id)

	// 从加密文件中加载私钥
	libp2pPrivateKeyFromFem, err := loadPrivateKey(id, "123456")
	assert.NoError(t, err)

	// 从原始数据中加载私钥
	libp2pPrivateKeyFromDer, err := libp2p_crypto.UnmarshalECDSAPrivateKey(privateKeyDer)
	assert.NoError(t, err)

	// 确认两者相同
	assert.Equal(t, libp2pPrivateKeyFromFem, libp2pPrivateKeyFromDer)
}

func TestLoadPrivateKey_NoPassword(t *testing.T) {

	privateKeyDer, id, err := generatePrivateKeyFile("")
	assert.NoError(t, err)
	fmt.Printf("id = %s\n", id)

	// 从加密文件中加载私钥
	libp2pPrivateKeyFromFem, err := loadPrivateKey(id, "")
	assert.NoError(t, err)

	// 从原始数据中加载私钥
	libp2pPrivateKeyFromDer, err := libp2p_crypto.UnmarshalECDSAPrivateKey(privateKeyDer)
	assert.NoError(t, err)

	// 确认两者相同
	assert.Equal(t, libp2pPrivateKeyFromFem, libp2pPrivateKeyFromDer)
}
