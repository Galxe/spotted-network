package signer_test

import (
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
)

func TestLocalSigner(t *testing.T) {
	// Create temporary directory for keystore files
	tmpDir, err := os.MkdirTemp("", "signer-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test key
	signingKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	// Create keystore and import key
	ks := keystore.NewKeyStore(tmpDir, keystore.StandardScryptN, keystore.StandardScryptP)
	password := "testpassword"

	// Import signing key
	signingAccount, err := ks.ImportECDSA(signingKey, password)
	require.NoError(t, err)

	// Get keystore path
	signingKeyPath := filepath.Join(tmpDir, signingAccount.URL.Path)

	// Create signer
	cfg := &signer.Config{
		SigningKeyPath: signingKeyPath,
		Password:       password,
	}
	s, err := signer.NewLocalSigner(cfg)
	require.NoError(t, err)

	t.Run("signing key functions", func(t *testing.T) {
		// Test signing address
		expectedSigningAddr := crypto.PubkeyToAddress(signingKey.PublicKey)
		assert.Equal(t, expectedSigningAddr, s.GetSigningAddress())

		// Test task response signing
		params := signer.TaskSignParams{
			User:        common.HexToAddress("0x1234567890123456789012345678901234567890"),
			ChainID:     1,
			BlockNumber: 12345,
			Key:         big.NewInt(1),
			Value:       big.NewInt(100),
		}

		// Sign task response
		sig, err := s.SignTaskResponse(params)
		require.NoError(t, err)

		// Verify task response
		err = s.VerifyTaskResponse(params, sig, s.GetSigningAddress().Hex())
		require.NoError(t, err)

		// Test invalid signer address
		wrongAddr := common.HexToAddress("0x0000000000000000000000000000000000000000")
		err = s.VerifyTaskResponse(params, sig, wrongAddr.Hex())
		require.Error(t, err)
	})

	t.Run("error cases", func(t *testing.T) {
		// Test with non-existent key file
		_, err := signer.NewLocalSigner(&signer.Config{
			SigningKeyPath: "non-existent-file",
			Password:       password,
		})
		require.Error(t, err)

		// Test with wrong password
		_, err = signer.NewLocalSigner(&signer.Config{
			SigningKeyPath: signingKeyPath,
			Password:       "wrongpassword",
		})
		require.Error(t, err)

		// Test signature verification with modified parameters
		params := signer.TaskSignParams{
			User:        common.HexToAddress("0x1234567890123456789012345678901234567890"),
			ChainID:     1,
			BlockNumber: 12345,
			Key:         big.NewInt(1),
			Value:       big.NewInt(100),
		}

		sig, err := s.SignTaskResponse(params)
		require.NoError(t, err)

		// Modify params and verify should fail
		params.Value = big.NewInt(200)
		err = s.VerifyTaskResponse(params, sig, s.GetSigningAddress().Hex())
		require.Error(t, err)
	})
}
