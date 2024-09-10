package account

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateAccount(t *testing.T) {
	_, _, err := NewAccount("蔡志宏", "13520746670", "123456")
	assert.NoError(t, err)
}
