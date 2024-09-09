package account

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateAccount(t *testing.T) {
	_, _, err := CreateAccount("蔡志宏", "13520746670", "abc123")
	assert.NoError(t, err)
}
