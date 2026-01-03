package security_test

import (
	"testing"
	"time"

	"github.com/modulix-systems/goose-talk/internal/config"
	"github.com/modulix-systems/goose-talk/internal/gateways/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecryptSymmetric(t *testing.T) {
	securityProvider := security.New(time.Hour, config.OTP_LENGTH, "Test App")
	plaintext := "Hello World. Lorem ipsum dolor sit amet"

	t.Run("success", func(t *testing.T) {
		key := securityProvider.GeneratePrivateKey()
		encrypted, err := securityProvider.EncryptSymmetric(plaintext, key)
		require.NoError(t, err)
		decrypted, err := securityProvider.DecryptSymmetric(encrypted, key)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("invalid decrypt key", func(t *testing.T) {
		key := securityProvider.GeneratePrivateKey()
		encrypted, err := securityProvider.EncryptSymmetric(plaintext, key)
		require.NoError(t, err)
		key = securityProvider.GeneratePrivateKey()
		decrypted, err := securityProvider.DecryptSymmetric(encrypted, key)
		assert.Error(t, err)
		assert.Empty(t, decrypted)
	})
}
