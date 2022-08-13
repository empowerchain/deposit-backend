package commons

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encore.dev/beta/errs"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/stretchr/testify/require"
	"testing"
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
			client:        ClientName,
			errorCode:     errs.OK,
		},
		{
			name:          "Different private and public key",
			signerPrivKey: secp256k1.GenPrivKey(),
			pubKey:        defaultSigner.PubKey().(*secp256k1.PubKey),
			client:        ClientName,
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
			authDataB64, pubKeyHex, err := GetToken(test.signerPrivKey, test.pubKey, test.client)
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
	privateKeyBytes, err := hex.DecodeString("2ea029a07b085d75ebfdef7318ec848ef0f6d8bef4506f363f433672f1e57f41")
	require.NoError(t, err)
	privateKey := &secp256k1.PrivKey{
		Key: privateKeyBytes,
	}
	publicKey := privateKey.PubKey().(*secp256k1.PubKey)

	token, pubKeyHex, err := GetToken(privateKey, publicKey, ClientName)
	require.NoError(t, err)

	uid, err := AuthHandler(context.Background(), token)
	require.NoError(t, err)
	require.Equal(t, pubKeyHex, string(uid))
}

func TestInvalidSignature(t *testing.T) {
	privateKeyBytes, err := hex.DecodeString("2ea029a07b085d75ebfdef7318ec848ef0f6d8bef4506f363f433672f1e57f41")
	require.NoError(t, err)
	privateKey := &secp256k1.PrivKey{
		Key: privateKeyBytes,
	}
	publicKey := privateKey.PubKey().(*secp256k1.PubKey)

	token, _, err := GetToken(privateKey, publicKey, ClientName)
	require.NoError(t, err)

	b, err := base64.StdEncoding.DecodeString(token)
	require.NoError(t, err)

	var authData AuthData
	err = json.Unmarshal(b, &authData)
	require.NoError(t, err)

	authData.Signature = "0307543f39d8887f493049c72526e4f943e853f417f487a6ae25e7cab9c89bb31cdb823c8986279b7a08fe7ebaf9037bd1e9163ef13b6b00b1835ac0c5d48f8b"
	b2, err := json.Marshal(authData)
	require.NoError(t, err)

	tokenWithInvalidSignature := base64.StdEncoding.EncodeToString(b2)
	_, err = AuthHandler(context.Background(), tokenWithInvalidSignature)
	require.EqualError(t, err, "unauthenticated: failed to verify signature")
}

func TestFrontendCreatedAuth(t *testing.T) {
	token := "eyJwYXlsb2FkIjoiZXlKd2RXSkxaWGtpT2lJd00yVmxPVGswTkRVd1ptWXlaVGt5WmpRNFpETmpObUZrTXpCbVpUSXhNRFJrWXpoaU1qVXhOREEyWWpFMVltSmxZVEZpWVRabE5UVXhOak5qWXpJMlpUa2lMQ0pqYkdsbGJuUWlPaUpsYlhCdmQyVnlMV1JsY0c5emFYUXRZWEJ3SWl3aWRHbHRaWE4wWVcxd0lqb3hOalUyT1RRek16QXhmUT09Iiwic2lnbmF0dXJlIjoiMDMwNzU0M2YzOWQ4ODg3ZjQ5MzA0OWM3MjUyNmU0Zjk0M2U4NTNmNDE3ZjQ4N2E2YWUyNWU3Y2FiOWM4OWJiMzFjZGI4MjNjODk4NjI3OWI3YTA4ZmU3ZWJhZjkwMzdiZDFlOTE2M2VmMTNiNmIwMGIxODM1YWMwYzVkNDhmOGIifQ=="
	uid, err := AuthHandler(context.Background(), token)
	require.NoError(t, err)
	require.Equal(t, "03ee994450ff2e92f48d3c6ad30fe2104dc8b251406b15bbea1ba6e55163cc26e9", string(uid))
}
