package commons

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encore.dev/beta/errs"
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var defaultSigner = secp256k1.GenPrivKey()

func TestAuthHandler(t *testing.T) {
	tests := []struct {
		name          string
		signerPrivKey *secp256k1.PrivKey
		pubKey        *secp256k1.PubKey
		client        string
		errorCode     errs.ErrCode
	}{
		{
			name:          "Happy path",
			signerPrivKey: defaultSigner,
			pubKey:        defaultSigner.PubKey().(*secp256k1.PubKey),
			client:        clientName,
			errorCode:     errs.OK,
		},
		{
			name:          "Different private and public key",
			signerPrivKey: secp256k1.GenPrivKey(),
			pubKey:        defaultSigner.PubKey().(*secp256k1.PubKey),
			client:        clientName,
			errorCode:     errs.Unauthenticated,
		},
		{
			name:          "Invalid client name",
			signerPrivKey: defaultSigner,
			pubKey:        defaultSigner.PubKey().(*secp256k1.PubKey),
			client:        "invalidClientName",
			errorCode:     errs.Unauthenticated,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			authDataB64, pubKeyHex, err := getToken(test.signerPrivKey, test.pubKey, test.client)
			require.NoError(t, err)
			uid, err := AuthHandler(context.Background(), authDataB64)

			if test.errorCode == errs.OK {
				require.NoError(t, err)
				require.Equal(t, string(uid), pubKeyHex)
			} else {
				require.Error(t, err)
				require.Equal(t, test.errorCode, err.(*errs.Error).Code)
			}
		})
	}
}

func TestFromPrivateKey(t *testing.T) {
	privateKeyBytes, err := hex.DecodeString("1eaa8d64b64c130d8690791e16e800f8e1bac55dd49aed88b325f3818770aa31")
	require.NoError(t, err)
	privateKey := &secp256k1.PrivKey{
		Key: privateKeyBytes,
	}
	publicKey := privateKey.PubKey().(*secp256k1.PubKey)

	token, pubKeyHex, err := getToken(privateKey, publicKey, clientName)
	require.NoError(t, err)

	uid, err := AuthHandler(context.Background(), token)
	require.NoError(t, err)
	require.Equal(t, pubKeyHex, string(uid))
}

func getToken(signerPrivKey *secp256k1.PrivKey, pubKey *secp256k1.PubKey, client string) (string, string, error) {
	pubKeyHex := hex.EncodeToString(pubKey.Bytes())

	payloadStr := fmt.Sprintf(`{
  "pubKey": "%s",
  "client":"%s",
  "timestamp":%d
}`, pubKeyHex, client, time.Now().Unix())
	payload := base64.StdEncoding.EncodeToString([]byte(payloadStr))

	payloadSignatureB, err := signerPrivKey.Sign([]byte(payload))
	if err != nil {
		return "", "", err
	}

	authData := fmt.Sprintf(`{
  "payload": "%s",
  "signature": "%s"
}`, payload, hex.EncodeToString(payloadSignatureB))

	authDataB64 := base64.StdEncoding.EncodeToString([]byte(authData))

	return authDataB64, pubKeyHex, nil
}

func TestFrontendCreatedAuth(t *testing.T) {
	token := "eyJwYXlsb2FkIjoiZXlKd2RXSkxaWGtpT2lJd00yVmxPVGswTkRVd1ptWXlaVGt5WmpRNFpETmpObUZrTXpCbVpUSXhNRFJrWXpoaU1qVXhOREEyWWpFMVltSmxZVEZpWVRabE5UVXhOak5qWXpJMlpUa2lMQ0pqYkdsbGJuUWlPaUpsYlhCdmQyVnlMV1JsY0c5emFYUXRZWEJ3SWl3aWRHbHRaWE4wWVcxd0lqb3hOalUyT1RRek16QXhmUT09Iiwic2lnbmF0dXJlIjoiMDMwNzU0M2YzOWQ4ODg3ZjQ5MzA0OWM3MjUyNmU0Zjk0M2U4NTNmNDE3ZjQ4N2E2YWUyNWU3Y2FiOWM4OWJiMzFjZGI4MjNjODk4NjI3OWI3YTA4ZmU3ZWJhZjkwMzdiZDFlOTE2M2VmMTNiNmIwMGIxODM1YWMwYzVkNDhmOGIifQ=="
	uid, err := AuthHandler(context.Background(), token)
	require.NoError(t, err)
	require.Equal(t, "03ee994450ff2e92f48d3c6ad30fe2104dc8b251406b15bbea1ba6e55163cc26e9", string(uid))
}
